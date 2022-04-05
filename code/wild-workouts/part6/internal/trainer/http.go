package main

import (
	"net/http"

	"github.com/MaksimDzhangirov/three-dots/code/wild-workouts/part6/internal/trainer/domain/hour"
	"github.com/MaksimDzhangirov/three-dots/part6/internal/common/auth"
	"github.com/MaksimDzhangirov/three-dots/part6/internal/common/server/httperr"
	"github.com/go-chi/render"
)

type HttpServer struct {
	db             db
	hourRepository hour.Repository
}

func (h HttpServer) GetTrainerAvailableHours(w http.ResponseWriter, r *http.Request, params GetTrainerAvailableHoursParams) {
	queryParams := r.Context().Value("GetTrainerAvailableHoursParams").(*GetTrainerAvailableHoursParams)

	if queryParams.DateFrom.After(queryParams.DateTo) {
		httperr.BadRequest("date-from-after-date-to", nil, w, r)
		return
	}

	dates, err := h.db.GetDates(r.Context(), queryParams)
	if err != nil {
		httperr.InternalError("unable-to-get-dates", err, w, r)
		return
	}

	render.Respond(w, r, dates)
}

func (h HttpServer) MakeHourAvailable(w http.ResponseWriter, r *http.Request) {
	user, err := auth.UserFromCtx(r.Context())
	if err != nil {
		httperr.Unauthorised("no-user-found", err, w, r)
		return
	}

	if user.Role != "trainer" {
		httperr.Unauthorised("invalid-role", nil, w, r)
		return
	}

	hourUpdate := &HourUpdate{}
	if err := render.Decode(r, hourUpdate); err != nil {
		httperr.BadRequest("unable-to-update-availability", err, w, r)
		return
	}

	for _, hourToUpdate := range hourUpdate.Hours {
		if err := h.hourRepository.UpdateHour(r.Context(), hourToUpdate, func(h *hour.Hour) (*hour.Hour, error) {
			if err := h.MakeAvailable(); err != nil {
				return nil, err
			}
			return h, nil
		}); err != nil {
			httperr.InternalError("unable-to-update-availability", err, w, r)
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h HttpServer) MakeHourUnavailable(w http.ResponseWriter, r *http.Request) {
	user, err := auth.UserFromCtx(r.Context())
	if err != nil {
		httperr.Unauthorised("no-user-found", err, w, r)
		return
	}
	if user.Role != "trainer" {
		httperr.Unauthorised("invalid-role", nil, w, r)
		return
	}

	hourUpdate := &HourUpdate{}
	if err := render.Decode(r, hourUpdate); err != nil {
		httperr.BadRequest("unable-to-update-availability", err, w, r)
		return
	}

	for _, hourToUpdate := range hourUpdate.Hours {
		if err := h.hourRepository.UpdateHour(r.Context(), hourToUpdate, func(h *hour.Hour) (*hour.Hour, error) {
			if err := h.MakeNotAvailable(); err != nil {
				return nil, err
			}
			return h, nil
		}); err != nil {
			httperr.InternalError("unable-to-update-availability", err, w, r)
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}
