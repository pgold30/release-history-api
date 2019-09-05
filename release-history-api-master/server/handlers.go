package server

import (
	"crypto/subtle"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type Handler struct {
	sqlDB *sql.DB
	user string
	password string
}

func NewHandler(sqlDB *sql.DB, user string, password string) *Handler {
	return &Handler{
		sqlDB:    sqlDB,
		user:     user,
		password: password,
	}
}

func (h *Handler) BasicAuth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()

		if !ok || subtle.ConstantTimeCompare([]byte(user), []byte(h.user)) != 1 || subtle.ConstantTimeCompare([]byte(pass), []byte(h.password)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="Please authenticate"`)
			w.WriteHeader(401)
			w.Write([]byte("Unauthorised.\n"))
			return
		}

		handler(w, r)
	}
}

func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) DeploymentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		h.newDeployment(w, r)
		return
	}
	if r.Method == "GET" {
		h.listDeployments(w, r)
		return
	}
	w.WriteHeader(http.StatusBadRequest)
	return
}

func (h *Handler) newDeployment(w http.ResponseWriter, r *http.Request) {
	deployment := &DeploymentInput{}
	err := json.NewDecoder(r.Body).Decode(deployment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if deployment.Project == "" || deployment.Service == "" || deployment.Tag == "" || deployment.Environment == "" {
		http.Error(w, errors.New("invalid deployment, missing required field").Error(), http.StatusBadRequest)
		return
	}

	response := DeploymentOutput{}
	err = h.sqlDB.QueryRow(`INSERT INTO deployment(d_project, d_service, d_tag, d_environment) VALUES($1, $2, $3, $4) RETURNING *`,
		deployment.Project, deployment.Service, deployment.Tag, deployment.Environment).Scan(&response.ID, &response.Project, &response.Service, &response.Environment, &response.Tag, &response.Date)
	if err != nil {
		fmt.Printf("Execute query error: %+v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.writeResponse(response, w)
}

func (h *Handler) listDeployments(w http.ResponseWriter, r *http.Request) {
	var err error
	// Parse date. Expected format = 2019-04-09T15:39:05.857844Z
	date := time.Now()
	dateStr, ok := r.URL.Query()["date"]
	if ok && len(dateStr) == 1 {
		date, err = time.Parse(time.RFC3339, dateStr[0])
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	environment := "prod"
	envStr, ok := r.URL.Query()["environment"]
	if ok && len(envStr) == 1 {
		environment = envStr[0]
	}

	project := "alt"
	projStr, ok := r.URL.Query()["project"]
	if ok && len(projStr) == 1 {
		project = projStr[0]
	}

	var query string

	if showAll, ok := r.URL.Query()["showAll"]; ok && len(projStr) == 1 && showAll[0] == "true" {
		query = `
			select * from deployment
			where d_date <= $1
			and d_environment = $2
			and d_project = $3;
		`
	} else {
		query = `
			select d1.* from deployment d1
			inner join (
				select d_service, d_environment, max(d_id) as d_id
				from deployment
				where d_date <= $1
				and d_environment = $2
				and d_project = $3
				group by 1, 2
			) d2 on d1.d_id = d2.d_id;
		`
	}
	deployments, err := h.findDeployments(query, date, environment, project)
	if err != nil {
		fmt.Printf("Unable to get deployments for date %s, environment %s and project %s: %+v\n", date, environment, project, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.writeResponse(deployments, w)
}

func (h *Handler) ReleaseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		h.newRelease(w, r)
		return
	}
	if r.Method == "GET" {
		h.listReleases(w, r)
		return
	}
	w.WriteHeader(http.StatusBadRequest)
	return
}

func (h *Handler) newRelease(w http.ResponseWriter, r *http.Request) {
	release := &ReleaseInput{}
	err := json.NewDecoder(r.Body).Decode(release)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if release.Project == "" || release.Number == "" {
		http.Error(w, errors.New("invalid release, missing required field").Error(), http.StatusBadRequest)
		return
	}

	response := ReleaseOutput{}
	err = h.sqlDB.QueryRow(`INSERT INTO release(r_project, r_number) VALUES($1, $2) RETURNING *`,
		release.Project, release.Number).Scan(&response.ID, &response.Project, &response.Number, &response.Date)
	if err != nil {
		fmt.Printf("Execute query error: %+v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = h.fetchDeploymentsForRelease(&response)
	if err != nil {
		fmt.Printf("Unable to get deployments for date %s, environment prod and project %s: %+v\n", response.Date, response.Project, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.writeResponse(response, w)
}

func (h *Handler) listReleases(w http.ResponseWriter, r *http.Request) {
	var err error
	// Parse date. Expected format = 2019-04-09T15:39:05.857844Z
	date := time.Now()
	dateStr, ok := r.URL.Query()["date"]
	if ok && len(dateStr) == 1 {
		date, err = time.Parse(time.RFC3339, dateStr[0])
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	project := "alt"
	projStr, ok := r.URL.Query()["project"]
	if ok && len(projStr) == 1 {
		project = projStr[0]
	}

	var rows *sql.Rows

	if showAll, ok := r.URL.Query()["showAll"]; ok && len(showAll) == 1 && showAll[0] == "true" {
		rows, err = h.sqlDB.Query(`
			select * from release
			where r_date <= $1
			and r_project = $2;
		`, date, project)
	} else if number, ok := r.URL.Query()["number"]; ok && len(number) == 1 && number[0] != "" {
		rows, err = h.sqlDB.Query(`
			select * from release
			where r_number = $1
			and r_project = $2;
		`, number[0], project)
	} else {
		rows, err = h.sqlDB.Query(`
			select r1.* from release r1
			inner join (
				select r_project, max(r_id) as r_id
				from release
				where r_date <= $1
				and r_project = $2
				group by 1
			) r2 on r1.r_id = r2.r_id;
		`, date, project)
	}

	if err != nil {
		fmt.Printf("Execute query error: %+v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	releases := []*ReleaseOutput{}
	for rows.Next() {
		var release ReleaseOutput
		err := rows.Scan(&release.ID, &release.Project, &release.Number, &release.Date)
		if err != nil {
			fmt.Printf("Execute query error: %+v\n", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		err = h.fetchDeploymentsForRelease(&release)
		if err != nil {
			fmt.Printf("Unable to get deployments for date %s, environment prod and project %s: %+v\n", release.Date, release.Project, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		releases = append(releases, &release)
	}
	h.writeResponse(releases, w)
}

func (h *Handler) fetchDeploymentsForRelease(release *ReleaseOutput) error {
	deployments, err := h.findDeployments(
		`
				select d1.* from deployment d1
				inner join (
					select d_service, d_environment, max(d_id) as d_id
					from deployment
					where d_date <= $1
					and d_environment = $2
					and d_project = $3
					group by 1, 2
				) d2 on d1.d_id = d2.d_id;
			`,
		release.Date, "prod", release.Project)
	if err != nil {
		return err
	}
	release.Deployments = deployments
	return nil
}

func (h *Handler) findDeployments(query string, date time.Time, environment string, project string) ([]*DeploymentOutput, error) {
	// Get latest deployment per environment
	rows, err := h.sqlDB.Query(query, date, environment, project)
	if err != nil {
		return nil, err
	}

	deployments := []*DeploymentOutput{}
	for rows.Next() {
		var deployment DeploymentOutput
		err := rows.Scan(&deployment.ID, &deployment.Project, &deployment.Service, &deployment.Environment, &deployment.Tag, &deployment.Date)
		if err != nil {
			return nil, err
		}
		deployments = append(deployments, &deployment)
	}
	return deployments, nil
}

func (h *Handler) writeResponse(response interface{}, w http.ResponseWriter) {
	b, err := json.Marshal(response)
	if err != nil {
		fmt.Printf("Unable to marshall response: %+v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	_, err = w.Write(b)
	if err != nil {
		fmt.Printf("Unable to write response: %+v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
