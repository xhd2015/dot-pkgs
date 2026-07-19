package iterm2

// TabStatusEntry is one tab's status within a set.
type TabStatusEntry struct {
	TabID string
	State string // "running" | "idle" | "missing" | "unknown"
}

// TabSetStatus is the aggregate status of a tab set.
type TabSetStatus struct {
	SetName    string
	WindowID   string
	WindowName string
	Warning    string
	Instances  int
	Tabs       []TabStatusEntry
}

// StatusTabSet reports per-config-tab state using Find + Busy from cfg.
//
//	busy → "running", idle → "idle", unknown → "unknown", absent → "missing"
func StatusTabSet(spec TabSetSpec, cfg *TabSetConfig) (*TabSetStatus, error) {
	cfg = normalizeTabSetConfig(cfg)
	refs, err := cfg.Find(spec.Name)
	if err != nil {
		return nil, err
	}

	windows := uniqueWindowIDs(refs)
	st := &TabSetStatus{
		SetName:   spec.Name,
		Instances: len(windows),
	}
	if len(windows) > 0 {
		st.WindowID = windows[0]
	}
	if len(windows) > 1 {
		st.Warning = "multiple windows host this tab set"
	}

	byTab := map[string]TabSessionRef{}
	for _, ref := range refs {
		if _, exists := byTab[ref.TabID]; !exists {
			byTab[ref.TabID] = ref
		}
	}

	for _, tab := range spec.Tabs {
		ref, ok := byTab[tab.ID]
		if !ok {
			st.Tabs = append(st.Tabs, TabStatusEntry{TabID: tab.ID, State: "missing"})
			continue
		}
		state := "unknown"
		switch cfg.Busy(ref) {
		case BusyStateBusy:
			state = "running"
		case BusyStateIdle:
			state = "idle"
		case BusyStateUnknown:
			state = "unknown"
		}
		st.Tabs = append(st.Tabs, TabStatusEntry{TabID: tab.ID, State: state})
	}
	return st, nil
}
