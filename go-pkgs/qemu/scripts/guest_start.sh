#!/bin/bash
# Idempotent qemu guest. Invoked as: sh -c "$(script)" -- <action>
set -euo pipefail
DIR="${QEMU_GUEST_DIR:-/root/qemu-guest}"
USER="${QEMU_GUEST_USER:-debian}"
PORT="${QEMU_GUEST_SSH_PORT:-22221}"
MEM="${QEMU_GUEST_MEM_MB:-1024}"
CPUS="${QEMU_GUEST_CPUS:-2}"
ACCEL="${QEMU_GUEST_ACCEL:-tcg}"
BACKING_NAME="${QEMU_BACKING_NAME:-debian-12-genericcloud-amd64.qcow2}"
IMAGE_URL="${QEMU_BACKING_URL:-https://cloud.debian.org/images/cloud/bookworm/latest/debian-12-genericcloud-amd64.qcow2}"
PKGVER="${QEMU_PKG_VERSION:-1:5.2+dfsg-11+deb11u3}"
PIDFILE=$DIR/qemu.pid
SERIAL=$DIR/serial.log
MONITOR=$DIR/monitor.sock
OVERLAY=$DIR/overlay.qcow2
BACKING=$DIR/$BACKING_NAME
KEY=$DIR/guest_ed25519
SEED=$DIR/seed.iso
ACTION=${1:-start}

ssh_cmd() {
	ssh -i "$KEY" -p "$PORT" \
		-o StrictHostKeyChecking=no \
		-o UserKnownHostsFile=/dev/null \
		-o IdentitiesOnly=yes \
		-o BatchMode=yes \
		-o ConnectTimeout=5 \
		-o LogLevel=ERROR \
		"$USER@127.0.0.1" "$@"
}

qemu_alive() {
	[[ -f "$PIDFILE" ]] && kill -0 "$(cat "$PIDFILE")" 2>/dev/null
}

qemu_accel_arg() {
	case "${ACCEL:-tcg}" in
	hvf) echo hvf ;;
	auto)
		if [[ "$(uname -s)" = Darwin ]]; then
			echo hvf
		else
			echo tcg,thread=multi
		fi
		;;
	tcg|*) echo tcg,thread=multi ;;
	esac
}

ensure_qemu_bin() {
	if command -v qemu-system-x86_64 >/dev/null 2>&1; then
		return 0
	fi
	export DEBIAN_FRONTEND=noninteractive
	apt-get update -y -o Acquire::Check-Valid-Until=false
	apt-get install -y --no-install-recommends \
		qemu-system-x86="$PKGVER" \
		qemu-system-common="$PKGVER" \
		qemu-system-data="$PKGVER" \
		qemu-utils="$PKGVER" \
		genisoimage ca-certificates || dpkg --configure -a
	command -v qemu-system-x86_64 >/dev/null
}

write_seed() {
	mkdir -p "$DIR/seed"
	PUB=$(cat "${KEY}.pub")
	cat >"$DIR/seed/user-data" <<EOF
#cloud-config
hostname: qemu-guest
manage_etc_hosts: true
users:
  - name: debian
    sudo: ALL=(ALL) NOPASSWD:ALL
    groups: sudo
    shell: /bin/bash
    lock_passwd: true
    ssh_authorized_keys:
      - $PUB
ssh_pwauth: false
package_update: false
package_upgrade: false
final_message: "CLOUD_INIT_DONE"
EOF
	cat >"$DIR/seed/meta-data" <<EOF
instance-id: qemu-guest-1
local-hostname: qemu-guest
EOF
	genisoimage -output "$SEED" -volid cidata -joliet -rock -quiet "$DIR/seed/user-data" "$DIR/seed/meta-data"
}

start_qemu() {
	local cdrom=()
	if [[ -f "$SEED" ]]; then
		cdrom=(-cdrom "$SEED")
	fi
	rm -f "$MONITOR"
	: >>"$SERIAL"
	qemu-system-x86_64 \
		-name qemu-guest \
		-accel "$(qemu_accel_arg)" \
		-cpu qemu64 \
		-smp "$CPUS" \
		-m "$MEM" \
		-machine pc \
		-display none \
		-daemonize \
		-pidfile "$PIDFILE" \
		-serial "file:$SERIAL" \
		-monitor "unix:$MONITOR,server,nowait" \
		-device virtio-rng-pci \
		-drive "if=virtio,file=$OVERLAY,format=qcow2,discard=unmap" \
		"${cdrom[@]}" \
		-netdev "user,id=n0,hostfwd=tcp:127.0.0.1:${PORT}-:22" \
		-device virtio-net-pci,netdev=n0
}

wait_ssh() {
	local i
	for i in $(seq 1 90); do
		if [[ -f "$KEY" ]] && ssh_cmd true 2>/dev/null; then
			echo guest_ssh=ok
			return 0
		fi
		sleep 2
	done
	echo guest_ssh=fail
	return 1
}

case "$ACTION" in
start)
	mkdir -p "$DIR"
	if qemu_alive; then
		echo skip=running
		echo qemu_pid=$(cat "$PIDFILE")
		if [[ -f "$KEY" ]] && ssh_cmd true 2>/dev/null; then
			echo guest_ssh=ok
		else
			echo guest_ssh=fail
		fi
		exit 0
	fi
	ensure_qemu_bin
	if [[ ! -s "$BACKING" ]]; then
		wget -O "$BACKING.partial" "$IMAGE_URL"
		mv "$BACKING.partial" "$BACKING"
	fi
	if [[ ! -f "$KEY" ]]; then
		ssh-keygen -t ed25519 -f "$KEY" -N '' -C qemu-guest
	fi
	NEW_OVERLAY=0
	if [[ ! -f "$OVERLAY" ]]; then
		qemu-img create -f qcow2 -F qcow2 -b "$BACKING" "$OVERLAY" 8G
		write_seed
		NEW_OVERLAY=1
	fi
	start_qemu
	echo qemu_pid=$(cat "$PIDFILE")
	if [[ "$NEW_OVERLAY" = 1 ]]; then
		wait_ssh
	elif [[ -f "$KEY" ]] && ssh_cmd true 2>/dev/null; then
		echo guest_ssh=ok
	else
		echo guest_ssh=wait
	fi
	;;
stop)
	if qemu_alive; then
		pid=$(cat "$PIDFILE")
		kill "$pid" 2>/dev/null || true
		sleep 1
		kill -9 "$pid" 2>/dev/null || true
		echo stopped pid=$pid
	else
		echo skip=already_stopped
	fi
	rm -f "$PIDFILE" "$MONITOR"
	;;
status)
	if qemu_alive; then
		echo qemu_alive=yes
		echo qemu_pid=$(cat "$PIDFILE")
	else
		echo qemu_alive=no
		echo qemu_pid=
	fi
	if command -v qemu-system-x86_64 >/dev/null 2>&1; then
		echo qemu_bin=$(command -v qemu-system-x86_64)
	else
		echo qemu_bin=missing
	fi
	if [[ -f "$OVERLAY" ]]; then
		echo overlay_bytes=$(stat -c %s "$OVERLAY" 2>/dev/null || echo 0)
	else
		echo overlay_bytes=
	fi
	if [[ -f "$BACKING" ]]; then
		echo backing_bytes=$(stat -c %s "$BACKING" 2>/dev/null || echo 0)
	else
		echo backing_bytes=
	fi
	if [[ -f "$KEY" ]] && ssh_cmd true 2>/dev/null; then
		echo guest_ssh=ok
	else
		echo guest_ssh=fail
	fi
	;;
purge)
	if qemu_alive; then
		pid=$(cat "$PIDFILE")
		kill "$pid" 2>/dev/null || true
		sleep 1
		kill -9 "$pid" 2>/dev/null || true
	fi
	rm -f "$PIDFILE" "$MONITOR" "$OVERLAY" "$SERIAL" "$SEED"
	if [[ "${2:-}" = "--deep" ]]; then
		rm -f "$BACKING" "$BACKING.partial"
	fi
	echo purged
	;;
logs)
	N=${2:-80}
	if [[ -f "$SERIAL" ]]; then
		tail -n "$N" "$SERIAL"
	else
		echo "no serial log"
	fi
	;;
*)
	echo "unknown action $ACTION" >&2
	exit 1
	;;
esac
