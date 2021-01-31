package domain

type Container struct {
	ID          string   `json:"id"`
	Names       []string `json:"names"`
	State       string   `json:"state"`
	Status      string   `json:"status"`
	PublicPorts []uint16 `json:"publicPorts"`
}
