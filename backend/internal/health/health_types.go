package health

// Liveness probe types

type AliveResponse struct {
	Status string `json:"status"`
}

// Readiness probe types

type ReadyResponse struct {
	Status      string            `json:"status"`
	Dependecies []DependecyStatus `json:"dependencies"`
}

type DependecyStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}
