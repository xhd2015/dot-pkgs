#!/bin/bash
# Guest cloudflared (quick tunnel) inside qemu.
# Invoked as: sh -c "$(script)" -- <action> [args]
set -euo pipefail
DIR="${QEMU_GUEST_DIR:-/root/qemu-guest}"
USER="${QEMU_GUEST_USER:-debian}"
PORT="${QEMU_GUEST_SSH_PORT:-22221}"
KEY=$DIR/guest_ed25519
QEMU_PID=$DIR/qemu.pid
CF_BIN="${QEMU_CF_BIN:-/usr/local/bin/cloudflared}"
CF_PID="${QEMU_CF_PID:-/var/run/work-qemu-cloudflared.pid}"
CF_LOG="${QEMU_CF_LOG:-/var/log/work-qemu-cloudflared.log}"
CF_LOGIN_LOG=/var/log/work-qemu-cloudflared-login.log
CF_URL_FILE=/root/.cloudflared/work-qemu-url
CERT="${QEMU_CF_CERT:-/root/.cloudflared/cert.pem}"
OLD_LOG=/var/log/cloudflared.log
CF_VER="${QEMU_CF_VERSION:-2025.11.1}"
CF_DL=https://github.com/cloudflare/cloudflared/releases/download/${CF_VER}/cloudflared-linux-amd64
DEFAULT_ORIGIN="${QEMU_CF_ORIGIN_URL:-http://127.0.0.1:8080}"
ACTION=${1:-status}
ORIGIN=${2:-$DEFAULT_ORIGIN}

ssh_base() {
	ssh -i "$KEY" -p "$PORT" \
		-o StrictHostKeyChecking=no \
		-o UserKnownHostsFile=/dev/null \
		-o IdentitiesOnly=yes \
		-o BatchMode=yes \
		-o ConnectTimeout=5 \
		-o LogLevel=ERROR \
		"$USER@127.0.0.1" "$@"
}

ssh_cmd() {
	ssh_base -- sudo -n "$@"
}

# Run a compound shell snippet as root in the guest.
# Must printf-%q the snippet: over ssh, `sudo bash -c "$1"` otherwise
# splits on spaces/semicolons so only the first word reaches -c
# (e.g. `export VAR=x; cloudflared …` becomes just `export`).
guest_sh() {
	ssh_base -- sudo -n bash -c "$(printf '%q' "$1")"
}

qemu_alive() {
	[[ -f "$QEMU_PID" ]] && kill -0 "$(cat "$QEMU_PID")" 2>/dev/null
}

guest_ok() {
	[[ -f "$KEY" ]] && ssh_cmd true 2>/dev/null
}

cf_pid() {
	ssh_cmd pgrep -x cloudflared 2>/dev/null | tr -d '\r' | awk 'NF{print; exit}'
}

login_pid() {
	guest_sh 'ps -eo pid,args | awk "/[c]loudflared .*tunnel login/ {print \$1; exit}"' 2>/dev/null | tr -d '\r' | tail -1
}

cert_ok() {
	ssh_cmd test -s "$CERT" 2>/dev/null
}

# Resolve tunnel UUID by name.
# 1) Buffer full `tunnel list` (do not pipe live cloudflared into awk).
# 2) Parse via here-string — `printf|awk;exit` under `set -o pipefail` still
#    yields SIGPIPE/exit 141 when awk stops early.
find_tunnel_uuid() {
	local name=$1
	local list uuid
	list=$(guest_sh "export TUNNEL_ORIGIN_CERT=$CERT; $CF_BIN tunnel list 2>/dev/null" || true)
	set +e
	uuid=$(awk -v n="$name" '$2==n{print $1; exit}' <<<"$list" | tr -d '\r')
	set -e
	printf '%s' "$uuid"
}

extract_url() {
	u=$(ssh_cmd cat "$CF_URL_FILE" 2>/dev/null | tr -d '\r' | awk 'NF{print; exit}')
	if [[ -n "${u:-}" ]]; then
		echo "$u"
		return 0
	fi
	guest_sh 'for f in /var/log/work-qemu-cloudflared.log /var/log/cloudflared.log; do
		[ -f "$f" ] || continue
		u=$(grep -oE "https://[a-z0-9.-]+\.(trycloudflare\.com|xhd2015\.xyz)" "$f" 2>/dev/null | tail -1)
		[ -n "$u" ] && echo "$u" && exit 0
	done
	exit 0' 2>/dev/null | tr -d '\r' | awk 'NF{print; exit}'
}

extract_login_url() {
	guest_sh 'u=$(grep -oE "https://dash\.cloudflare\.com[^[:space:]]+" /var/log/work-qemu-cloudflared-login.log 2>/dev/null | tail -1)
		[ -n "$u" ] && echo "$u" && exit 0
		grep -oE "https://[^[:space:]]+" /var/log/work-qemu-cloudflared-login.log 2>/dev/null | grep -v trycloudflare | tail -1
		exit 0' 2>/dev/null | tr -d '\r' | tail -1
}

print_status_keys() {
	if qemu_alive; then
		echo qemu_alive=yes
		echo qemu_pid=$(cat "$QEMU_PID")
	else
		echo qemu_alive=no
		echo qemu_pid=
	fi
	if guest_ok; then
		echo guest_ssh=ok
	else
		echo guest_ssh=fail
	fi
	echo origin=$ORIGIN
	if ! qemu_alive || ! guest_ok; then
		echo cf_alive=no
		echo cf_pid=
		echo cf_bin=missing
		echo url=
		echo cert=missing
		return 0
	fi
	if ssh_cmd test -x "$CF_BIN" 2>/dev/null; then
		echo cf_bin=ok
	else
		echo cf_bin=missing
	fi
	pid=$(cf_pid || true)
	if [[ -n "${pid:-}" ]]; then
		echo cf_alive=yes
		echo cf_pid=$pid
	else
		echo cf_alive=no
		echo cf_pid=
	fi
	u=$(extract_url || true)
	echo url=${u:-}
	if cert_ok; then
		echo cert=ok
	else
		echo cert=missing
	fi
}

ensure_cf() {
	if ssh_cmd test -x "$CF_BIN" 2>/dev/null; then
		echo cf_bin=ok
		return 0
	fi
	guest_sh "set -e
DL='$CF_DL'
tmp=/tmp/cloudflared.\$\$
if command -v curl >/dev/null 2>&1; then
  curl -fsSL -o \"\$tmp\" \"\$DL\"
elif command -v wget >/dev/null 2>&1; then
  wget -q -O \"\$tmp\" \"\$DL\"
else
  export DEBIAN_FRONTEND=noninteractive
  apt-get update -y -o Acquire::Check-Valid-Until=false
  apt-get install -y --no-install-recommends curl ca-certificates
  curl -fsSL -o \"\$tmp\" \"\$DL\"
fi
chmod +x \"\$tmp\"
mv \"\$tmp\" $CF_BIN
$CF_BIN --version >/dev/null
"
	echo cf_bin=ok
}

require_guest() {
	if ! qemu_alive; then
		print_status_keys
		exit 1
	fi
	if ! guest_ok; then
		print_status_keys
		exit 1
	fi
}

case "$ACTION" in
status)
	print_status_keys
	;;
start)
	require_guest
	ensure_cf
	pid=$(cf_pid || true)
	if [[ -n "${pid:-}" ]]; then
		echo skip=running
		print_status_keys
		exit 0
	fi
	guest_sh "set -e
mkdir -p /root/.cloudflared /var/log /var/run
: >> $CF_LOG
nohup $CF_BIN tunnel --no-autoupdate --url '$ORIGIN' >$CF_LOG 2>&1 &
echo \$! > $CF_PID
"
	url=
	i=0
	while [[ $i -lt 30 ]]; do
		url=$(extract_url || true)
		if [[ -n "${url:-}" ]]; then
			break
		fi
		pid=$(cf_pid || true)
		if [[ -z "${pid:-}" && $i -gt 3 ]]; then
			break
		fi
		sleep 2
		i=$((i + 1))
	done
	if [[ -n "${url:-}" ]]; then
		guest_sh "echo '$url' > $CF_URL_FILE" || true
	fi
	print_status_keys
	pid=$(cf_pid || true)
	if [[ -z "${pid:-}" ]]; then
		exit 1
	fi
	;;
stop)
	if ! qemu_alive || ! guest_ok; then
		echo skip=already_stopped
		exit 0
	fi
	pid=$(cf_pid || true)
	if [[ -n "${pid:-}" ]]; then
		ssh_cmd kill "$pid" 2>/dev/null || true
		sleep 1
		ssh_cmd kill -9 "$pid" 2>/dev/null || true
		echo stopped pid=$pid
	else
		echo skip=already_stopped
	fi
	ssh_cmd rm -f "$CF_PID" 2>/dev/null || true
	;;
logs)
	N=${2:-80}
	require_guest
	guest_sh "if [ -f $CF_LOG ]; then tail -n $N $CF_LOG
elif [ -f $OLD_LOG ]; then tail -n $N $OLD_LOG
else echo no cloudflared log
fi"
	;;
purge)
	if qemu_alive && guest_ok; then
		pid=$(cf_pid || true)
		if [[ -n "${pid:-}" ]]; then
			ssh_cmd kill "$pid" 2>/dev/null || true
			sleep 1
			ssh_cmd kill -9 "$pid" 2>/dev/null || true
		fi
		lp=$(login_pid || true)
		if [[ -n "${lp:-}" ]]; then
			ssh_cmd kill "$lp" 2>/dev/null || true
			ssh_cmd kill -9 "$lp" 2>/dev/null || true
		fi
		ssh_cmd rm -f "$CF_PID" "$CF_LOG" "$CF_LOGIN_LOG" "$CF_URL_FILE" "$OLD_LOG" 2>/dev/null || true
		if [[ "${2:-}" = "--deep" ]]; then
			ssh_cmd rm -f "$CERT" 2>/dev/null || true
			echo purged_cert=yes
		fi
	fi
	echo purged
	;;
login)
	require_guest
	ensure_cf
	if cert_ok; then
		echo skip=already_logged_in
		echo cert=ok
		print_status_keys
		exit 0
	fi
	lp=$(login_pid || true)
	if [[ -z "${lp:-}" ]]; then
		guest_sh "set -e
mkdir -p /root/.cloudflared /var/log
: >> $CF_LOGIN_LOG
nohup $CF_BIN tunnel login >$CF_LOGIN_LOG 2>&1 &
"
	fi
	i=0
	login_url=
	while [[ $i -lt 80 ]]; do
		if [[ -z "${login_url:-}" ]]; then
			login_url=$(extract_login_url || true)
			if [[ -n "${login_url:-}" ]]; then
				echo login_url=$login_url
			fi
		fi
		if cert_ok; then
			echo cert=ok
			print_status_keys
			exit 0
		fi
		sleep 2
		i=$((i + 1))
	done
	if [[ -n "${login_url:-}" ]]; then
		echo login_url=$login_url
	fi
	echo cert=fail
	print_status_keys
	exit 1
	;;
list)
	require_guest
	if ! cert_ok; then
		echo cert=missing
		exit 1
	fi
	guest_sh "export TUNNEL_ORIGIN_CERT=$CERT; $CF_BIN tunnel list"
	;;
named-start)
	NAME=${2:-}
	ORIGIN=${3:-$DEFAULT_ORIGIN}
	shift 3 || true
	require_guest
	ensure_cf
	if ! cert_ok; then
		echo cert=missing
		exit 1
	fi
	if [[ -z "$NAME" ]]; then
		echo "named-start requires tunnel name" >&2
		exit 1
	fi
	HOSTS=("$@")
	if [[ ${#HOSTS[@]} -eq 0 ]]; then
		echo "named-start requires at least one hostname" >&2
		exit 1
	fi
	UUID=$(find_tunnel_uuid "$NAME")
	if [[ -z "${UUID:-}" ]]; then
		guest_sh "export TUNNEL_ORIGIN_CERT=$CERT; $CF_BIN tunnel create $NAME"
		UUID=$(find_tunnel_uuid "$NAME")
	fi
	test -n "$UUID"
	echo uuid=$UUID
	CFG=/root/.cloudflared/${NAME}.yml
	cfg_printf="printf '%s\\n' 'tunnel: ${UUID}' 'credentials-file: /root/.cloudflared/${UUID}.json' 'protocol: http2' 'ingress:'"
	for h in "${HOSTS[@]}"; do
		cfg_printf="${cfg_printf} '  - hostname: ${h}' '    service: ${ORIGIN}'"
	done
	cfg_printf="${cfg_printf} '  - service: http_status:404' > ${CFG}"
	guest_sh "$cfg_printf"
	for h in "${HOSTS[@]}"; do
		guest_sh "export TUNNEL_ORIGIN_CERT=$CERT; $CF_BIN tunnel route dns -f $NAME $h" || true
		echo routed=$h
	done
	guest_sh "pkill -x cloudflared || true; sleep 1; pkill -9 -x cloudflared || true" || true
	sleep 1
	guest_sh "set -e
mkdir -p /root/.cloudflared /var/log /var/run
: >> $CF_LOG
nohup $CF_BIN --no-autoupdate tunnel --config $CFG run >>$CF_LOG 2>&1 &
echo \$! > $CF_PID
echo https://${HOSTS[0]} > $CF_URL_FILE
"
	sleep 2
	print_status_keys
	pid=$(cf_pid || true)
	if [[ -z "${pid:-}" ]]; then
		exit 1
	fi
	echo named=yes
	echo tunnel=$NAME
	;;
named-drop-canonical)
	NAME=${2:-}
	ORIGIN=${3:-$DEFAULT_ORIGIN}
	DROP=${4:-}
	KEEP=${5:-}
	require_guest
	if [[ -z "$NAME" || -z "$KEEP" ]]; then
		echo "named-drop-canonical requires name and keep hostname" >&2
		exit 1
	fi
	UUID=$(find_tunnel_uuid "$NAME")
	test -n "$UUID"
	CFG=/root/.cloudflared/${NAME}.yml
	guest_sh "printf '%s\\n' 'tunnel: ${UUID}' 'credentials-file: /root/.cloudflared/${UUID}.json' 'protocol: http2' 'ingress:' '  - hostname: ${KEEP}' '    service: ${ORIGIN}' '  - service: http_status:404' > ${CFG}"
	echo dropped=${DROP:-}
	echo kept=$KEEP
	guest_sh "pkill -x cloudflared || true; sleep 1; pkill -9 -x cloudflared || true" || true
	sleep 1
	guest_sh "set -e
: >> $CF_LOG
nohup $CF_BIN --no-autoupdate tunnel --config $CFG run >>$CF_LOG 2>&1 &
echo \$! > $CF_PID
echo https://$KEEP > $CF_URL_FILE
"
	sleep 2
	print_status_keys
	echo named=yes
	;;
named-apply)
	# named-apply NAME host1=origin1 host2=origin2 ...
	# Per-host origins; rewrites full ingress and restarts guest cloudflared only.
	NAME=${2:-}
	shift 2 || true
	require_guest
	ensure_cf
	if ! cert_ok; then
		echo cert=missing
		exit 1
	fi
	if [[ -z "$NAME" ]]; then
		echo "named-apply requires tunnel name" >&2
		exit 1
	fi
	if [[ $# -lt 1 ]]; then
		echo "named-apply requires at least one host=origin pair" >&2
		exit 1
	fi
	UUID=$(find_tunnel_uuid "$NAME")
	if [[ -z "${UUID:-}" ]]; then
		guest_sh "export TUNNEL_ORIGIN_CERT=$CERT; $CF_BIN tunnel create $NAME"
		UUID=$(find_tunnel_uuid "$NAME")
	fi
	test -n "$UUID"
	echo uuid=$UUID
	CFG=/root/.cloudflared/${NAME}.yml
	# Build ingress lines in a temp script on the guest
	INGRESS_BUILD="printf '%s\\n' 'tunnel: ${UUID}' 'credentials-file: /root/.cloudflared/${UUID}.json' 'protocol: http2' 'ingress:'"
	FIRST_HOST=
	for pair in "$@"; do
		host=${pair%%=*}
		origin=${pair#*=}
		if [[ -z "$host" || -z "$origin" || "$host" = "$pair" ]]; then
			echo "invalid pair $pair (want host=origin)" >&2
			exit 1
		fi
		if [[ -z "$FIRST_HOST" ]]; then
			FIRST_HOST=$host
		fi
		INGRESS_BUILD="${INGRESS_BUILD} '  - hostname: ${host}' '    service: ${origin}'"
		guest_sh "export TUNNEL_ORIGIN_CERT=$CERT; $CF_BIN tunnel route dns -f $NAME $host" || true
		echo routed=$host
		echo origin_$host=$origin
	done
	INGRESS_BUILD="${INGRESS_BUILD} '  - service: http_status:404' > ${CFG}"
	guest_sh "$INGRESS_BUILD"
	guest_sh "pkill -x cloudflared || true; sleep 1; pkill -9 -x cloudflared || true" || true
	sleep 1
	guest_sh "set -e
mkdir -p /root/.cloudflared /var/log /var/run
: >> $CF_LOG
nohup $CF_BIN --no-autoupdate tunnel --config $CFG run >>$CF_LOG 2>&1 &
echo \$! > $CF_PID
echo https://${FIRST_HOST} > $CF_URL_FILE
"
	pid=
	i=0
	while [[ $i -lt 15 ]]; do
		pid=$(cf_pid || true)
		if [[ -n "${pid:-}" ]]; then
			break
		fi
		sleep 1
		i=$((i + 1))
	done
	print_status_keys || true
	if [[ -z "${pid:-}" ]]; then
		echo "named-apply: cloudflared did not start" >&2
		guest_sh "tail -n 40 $CF_LOG" >&2 || true
		exit 1
	fi
	echo named=yes
	echo tunnel=$NAME
	echo applied=yes
	;;
*)
	echo "unknown action $ACTION" >&2
	exit 1
	;;
esac
