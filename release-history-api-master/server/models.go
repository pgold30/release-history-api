package server

import "time"

// DeploymentInput represents a new deployment to be created
type DeploymentInput struct {
	Project     string `json:"project"`
	Service     string `json:"service"`
	Environment string `json:"environment"`
	Tag         string `json:"tag"`
}

// DeploymentOutput represents a persisted deployment
type DeploymentOutput struct {
	ID          int64     `json:"id"`
	Project     string    `json:"project"`
	Service     string    `json:"service"`
	Environment string    `json:"environment"`
	Tag         string    `json:"tag"`
	Date        time.Time `json:"date"`
}

// ReleaseInput represents a new release to be created
type ReleaseInput struct {
	Project     string `json:"project"`
	Number      string `json:"number"`
}

// ReleaseOutput represents a persisted release
type ReleaseOutput struct {
	ID          int64               `json:"id"`
	Project     string              `json:"project"`
	Number      string              `json:"number"`
	Deployments []*DeploymentOutput `json:"deployments"`
	Date        time.Time           `json:"date"`
}
