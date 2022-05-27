package main

import (
	"encoding/json"
	"net/http"
)

func (app *application) healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	//js := `{"status": "available", "environment": %q, "version": %q}`
	//js = fmt.Sprintf(js, app.config.env, version)
	//
	//w.Header().Set("Content-Type", "application/json")
	//
	//w.Write([]byte(js))

	env := envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": app.config.env,
			"version":     version,
		},
	}

	err := app.writeJSON(w, http.StatusOK, env, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) exampleHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]string{
		"hello": "world",
	}

	w.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		app.logger.PrintError(err, nil)
		http.Error(w, "The server encountered a problem and could not process your request", http.StatusInternalServerError)
	}

}
