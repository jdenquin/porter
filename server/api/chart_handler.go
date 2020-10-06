package api

import (
	"encoding/json"
	"net/http"

	"github.com/porter-dev/porter/internal/forms"
)

// Enumeration of chart API error codes, represented as int64
const (
	ErrChartDecode ErrorCode = iota + 600
	ErrChartValidateFields
)

// HandleListCharts retrieves a list of charts with various filter options
func (app *App) HandleListCharts(w http.ResponseWriter, r *http.Request) {
	// get the filter options
	form := &forms.ListChartForm{}

	// decode from JSON to form value
	if err := json.NewDecoder(r.Body).Decode(form); err != nil {
		app.handleErrorFormDecoding(err, ErrChartDecode, w)
		return
	}

	form.PopulateHelmOptions(app.repo.User)

	// validate the form
	if err := app.validator.Struct(form); err != nil {
		app.handleErrorFormValidation(err, ErrChartValidateFields, w)
		return
	}

	// create a new agent
	agent, err := form.HelmOptions.ToAgent(app.logger, app.helmConf, app.HelmTestStorageDriver)

	releases, err := agent.ListReleases(form.HelmOptions.Namespace, form.ListFilter)

	if err != nil {
		app.handleErrorFormValidation(err, ErrChartValidateFields, w)
		return
	}

	if err := json.NewEncoder(w).Encode(releases); err != nil {
		app.handleErrorFormDecoding(err, ErrChartDecode, w)
		return
	}
}
