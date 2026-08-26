// Package dashboard aggregates admin KPI stats from the other services and
// serves them over a single endpoint + an SSE stream for a realtime admin
// dashboard.
package dashboard

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	myauth "comics-galore/backend/auth"
	mybilling "comics-galore/backend/billing"
	mycomics "comics-galore/backend/comics"
	myreading "comics-galore/backend/reading"
	myupload "comics-galore/backend/upload"

	"encore.dev/beta/auth"
	"encore.dev/beta/errs"
)

type DashboardResponse struct {
	Users         *myauth.DashboardStats           `json:"users"`
	Comics        *mycomics.ComicsStats            `json:"comics"`
	Billing       *mybilling.BillingStats          `json:"billing"`
	Reading       *myreading.ReadingStats          `json:"reading"`
	Storage       *myupload.StorageStats           `json:"storage"`
	DownloadTrend []myreading.DownloadTrendPoint   `json:"download_trend"`
	SignupTrend   []myauth.SignupTrendPoint        `json:"signup_trend"`
}

func aggregate(ctx context.Context) *DashboardResponse {
	resp := &DashboardResponse{
		DownloadTrend: []myreading.DownloadTrendPoint{},
		SignupTrend:   []myauth.SignupTrendPoint{},
	}

	if v, err := myauth.AdminDashboardStats(ctx); err == nil && v != nil {
		resp.Users = v
	}
	if v, err := myauth.GetSignupTrend(ctx); err == nil && v != nil && v.Points != nil {
		resp.SignupTrend = v.Points
	}
	if v, err := mycomics.GetComicsStats(ctx); err == nil && v != nil {
		resp.Comics = v
	}
	if v, err := mybilling.GetBillingStats(ctx); err == nil && v != nil {
		resp.Billing = v
	}
	if v, err := myreading.GetReadingStats(ctx); err == nil && v != nil {
		resp.Reading = v
	}
	if v, err := myupload.GetStorageStats(ctx); err == nil && v != nil {
		resp.Storage = v
	}
	if v, err := myreading.GetDownloadTrend(ctx); err == nil && v != nil && v.Points != nil {
		resp.DownloadTrend = v.Points
	}

	return resp
}

//encore:api auth method=GET path=/admin/dashboard
func GetDashboard(ctx context.Context) (*DashboardResponse, error) {
	ad := auth.Data().(*myauth.AuthData)
	if ad.Role != "admin" {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "admin only"}
	}
	return aggregate(ctx), nil
}

//encore:api auth raw method=GET path=/admin/dashboard-stream
func DashboardStream(w http.ResponseWriter, req *http.Request) {
	ad := auth.Data().(*myauth.AuthData)
	if ad.Role != "admin" {
		http.Error(w, "admin only", http.StatusForbidden)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	push := func() {
		data, err := json.Marshal(aggregate(req.Context()))
		if err != nil {
			return
		}
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
	}

	push()

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			push()
		case <-req.Context().Done():
			return
		}
	}
}
