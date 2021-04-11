package japi

/*
NodeStatus describes mesh node
*/
type NodeStatus struct {
	Synced        bool   `json:"synced"`
	SyncedLayer   uint64 `json:"syncedLayer,string"`
	CurrentLayer  uint64 `json:"currentLayer,string"`
	VerifiedLayer uint64 `json:"verifiedLayer,string"`
	Peers         uint64 `json:"peers,string"`
	MinPeers      uint64 `json:"minPeers,string"`
	MaxPeers      uint64 `json:"maxPeers,string"`
}

