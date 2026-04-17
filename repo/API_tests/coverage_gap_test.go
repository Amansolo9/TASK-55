package API_tests

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"clubops_portal/internal/services"
)

// --- GET / (dashboard, now secured) ---

func TestDashboardRedirectsUnauthenticated(t *testing.T) {
	app, st := setupApp(t)
	defer st.Close()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 302 && resp.StatusCode != 401 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected redirect/401 for unauthenticated dashboard, got %d body=%s", resp.StatusCode, string(body))
	}
}

func TestDashboardRendersForAuthenticatedAdmin(t *testing.T) {
	app, st := setupApp(t)
	defer st.Close()
	adminHash, _ := services.NewAuthService(st, 30*time.Minute, 5, 15*time.Minute).HashPassword("StrongAdmin123!")
	if err := st.UpdatePassword(1, adminHash, false); err != nil {
		t.Fatal(err)
	}
	auth := login(t, app, "admin", "StrongAdmin123!")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	addAuth(req, auth)
	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200 dashboard, got %d body=%s", resp.StatusCode, string(body))
	}
}

// --- GET /clubs/recruiting (public) ---

func TestPublicRecruitingPageRenders(t *testing.T) {
	app, st := setupApp(t)
	defer st.Close()
	req := httptest.NewRequest(http.MethodGet, "/clubs/recruiting", nil)
	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200 public recruiting page, got %d body=%s", resp.StatusCode, string(body))
	}
}

// --- GET /credits ---

func TestCreditsPageAdminAccess(t *testing.T) {
	app, st := setupApp(t)
	defer st.Close()
	adminHash, _ := services.NewAuthService(st, 30*time.Minute, 5, 15*time.Minute).HashPassword("StrongAdmin123!")
	if err := st.UpdatePassword(1, adminHash, false); err != nil {
		t.Fatal(err)
	}
	auth := login(t, app, "admin", "StrongAdmin123!")
	req := httptest.NewRequest(http.MethodGet, "/credits", nil)
	addAuth(req, auth)
	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200 credits page, got %d body=%s", resp.StatusCode, string(body))
	}
}

// --- GET /regions ---

func TestRegionsPageAdminAccess(t *testing.T) {
	app, st := setupApp(t)
	defer st.Close()
	adminHash, _ := services.NewAuthService(st, 30*time.Minute, 5, 15*time.Minute).HashPassword("StrongAdmin123!")
	if err := st.UpdatePassword(1, adminHash, false); err != nil {
		t.Fatal(err)
	}
	auth := login(t, app, "admin", "StrongAdmin123!")
	req := httptest.NewRequest(http.MethodGet, "/regions", nil)
	addAuth(req, auth)
	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200 regions page, got %d body=%s", resp.StatusCode, string(body))
	}
}

// --- GET /mdm ---

func TestMDMPageAdminAccess(t *testing.T) {
	app, st := setupApp(t)
	defer st.Close()
	adminHash, _ := services.NewAuthService(st, 30*time.Minute, 5, 15*time.Minute).HashPassword("StrongAdmin123!")
	if err := st.UpdatePassword(1, adminHash, false); err != nil {
		t.Fatal(err)
	}
	auth := login(t, app, "admin", "StrongAdmin123!")
	req := httptest.NewRequest(http.MethodGet, "/mdm", nil)
	addAuth(req, auth)
	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200 mdm page, got %d body=%s", resp.StatusCode, string(body))
	}
}

// --- GET /partials/budgets/changes ---

func TestBudgetChangesPartialRendersForOrganizer(t *testing.T) {
	app, st := setupApp(t)
	defer st.Close()
	authSvc := services.NewAuthService(st, 30*time.Minute, 5, 15*time.Minute)
	hash, _ := authSvc.HashPassword("OrganizerPass123!")
	if err := st.CreateUser("org-changes", hash, "organizer", int64Ptr(1)); err != nil {
		t.Fatal(err)
	}
	auth := login(t, app, "org-changes", "OrganizerPass123!")
	req := httptest.NewRequest(http.MethodGet, "/partials/budgets/changes", nil)
	addAuth(req, auth)
	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200 budget changes partial, got %d body=%s", resp.StatusCode, string(body))
	}
}

// --- POST /api/budgets/:id/change ---

func TestRequestBudgetChangeSmallPctUpdatesDirectly(t *testing.T) {
	app, st := setupApp(t)
	defer st.Close()
	authSvc := services.NewAuthService(st, 30*time.Minute, 5, 15*time.Minute)
	hash, _ := authSvc.HashPassword("OrganizerPass123!")
	if err := st.CreateUser("org-req-change", hash, "organizer", int64Ptr(1)); err != nil {
		t.Fatal(err)
	}
	if _, err := st.DB.Exec(`INSERT INTO budgets (id, club_id, account_code, campus_code, project_code, period_type, period_start, amount, spent, created_by, status) VALUES (30, 1, 'acct', 'camp', 'proj', 'monthly', '2026-03', 1000, 0, 1, 'active')`); err != nil {
		t.Fatal(err)
	}
	auth := login(t, app, "org-req-change", "OrganizerPass123!")
	form := url.Values{}
	form.Set("proposed_amount", "1050") // 5% change
	form.Set("reason", "small adjust")
	req := httptest.NewRequest(http.MethodPost, "/api/budgets/30/change", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	addAuth(req, auth)
	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200 on small change, got %d body=%s", resp.StatusCode, string(body))
	}
	var amount float64
	if err := st.DB.QueryRow(`SELECT amount FROM budgets WHERE id = 30`).Scan(&amount); err != nil {
		t.Fatal(err)
	}
	if amount != 1050 {
		t.Fatalf("expected amount 1050, got %v", amount)
	}
}

func TestRequestBudgetChangeLargePctCreatesPending(t *testing.T) {
	app, st := setupApp(t)
	defer st.Close()
	authSvc := services.NewAuthService(st, 30*time.Minute, 5, 15*time.Minute)
	hash, _ := authSvc.HashPassword("OrganizerPass123!")
	if err := st.CreateUser("org-req-big", hash, "organizer", int64Ptr(1)); err != nil {
		t.Fatal(err)
	}
	if _, err := st.DB.Exec(`INSERT INTO budgets (id, club_id, account_code, campus_code, project_code, period_type, period_start, amount, spent, created_by, status) VALUES (31, 1, 'acct', 'camp', 'proj', 'monthly', '2026-03', 1000, 0, 1, 'active')`); err != nil {
		t.Fatal(err)
	}
	auth := login(t, app, "org-req-big", "OrganizerPass123!")
	form := url.Values{}
	form.Set("proposed_amount", "1500") // 50% change
	form.Set("reason", "big expansion")
	req := httptest.NewRequest(http.MethodPost, "/api/budgets/31/change", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	addAuth(req, auth)
	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200 creating change request, got %d body=%s", resp.StatusCode, string(body))
	}
	var count int
	if err := st.DB.QueryRow(`SELECT COUNT(1) FROM budget_change_requests WHERE budget_id = 31 AND status = 'pending'`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("expected 1 pending change request, got %d", count)
	}
}

// --- POST /api/budgets/:id/spend ---

func TestRecordBudgetSpend(t *testing.T) {
	app, st := setupApp(t)
	defer st.Close()
	authSvc := services.NewAuthService(st, 30*time.Minute, 5, 15*time.Minute)
	hash, _ := authSvc.HashPassword("OrganizerPass123!")
	if err := st.CreateUser("org-spend", hash, "organizer", int64Ptr(1)); err != nil {
		t.Fatal(err)
	}
	if _, err := st.DB.Exec(`INSERT INTO budgets (id, club_id, account_code, campus_code, project_code, period_type, period_start, amount, spent, created_by, status) VALUES (40, 1, 'acct', 'camp', 'proj', 'monthly', '2026-03', 2000, 0, 1, 'active')`); err != nil {
		t.Fatal(err)
	}
	auth := login(t, app, "org-spend", "OrganizerPass123!")
	form := url.Values{}
	form.Set("spent", "750")
	req := httptest.NewRequest(http.MethodPost, "/api/budgets/40/spend", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	addAuth(req, auth)
	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200 recording spend, got %d body=%s", resp.StatusCode, string(body))
	}
	var spent float64
	if err := st.DB.QueryRow(`SELECT spent FROM budgets WHERE id = 40`).Scan(&spent); err != nil {
		t.Fatal(err)
	}
	if spent != 750 {
		t.Fatalf("expected spent=750, got %v", spent)
	}
}

func TestRecordBudgetSpendRejectsNegative(t *testing.T) {
	app, st := setupApp(t)
	defer st.Close()
	adminHash, _ := services.NewAuthService(st, 30*time.Minute, 5, 15*time.Minute).HashPassword("StrongAdmin123!")
	if err := st.UpdatePassword(1, adminHash, false); err != nil {
		t.Fatal(err)
	}
	if _, err := st.DB.Exec(`INSERT INTO budgets (id, club_id, account_code, campus_code, project_code, period_type, period_start, amount, spent, created_by, status) VALUES (41, 1, 'acct', 'camp', 'proj', 'monthly', '2026-03', 2000, 0, 1, 'active')`); err != nil {
		t.Fatal(err)
	}
	auth := login(t, app, "admin", "StrongAdmin123!")
	form := url.Values{}
	form.Set("spent", "-10")
	req := httptest.NewRequest(http.MethodPost, "/api/budgets/41/spend", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	addAuth(req, auth)
	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode == 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected non-200 for negative spend, got %d body=%s", resp.StatusCode, string(body))
	}
}

// --- POST /api/reviews/:id/appeal ---

func TestAppealReviewByAuthor(t *testing.T) {
	app, st := setupApp(t)
	defer st.Close()
	authSvc := services.NewAuthService(st, 30*time.Minute, 5, 15*time.Minute)
	memberHash, _ := authSvc.HashPassword("MemberPass123!")
	if err := st.CreateUser("member-appeal", memberHash, "member", nil); err != nil {
		t.Fatal(err)
	}
	member, err := st.FindUserByUsername("member-appeal")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.DB.Exec(`INSERT INTO reviews (id, club_id, fulfilled_order_id, site_id, member_id, reviewer_id, stars, comment, tags, appeal_status, hidden_reason, hidden_at, created_at) VALUES (50, 1, NULL, 1, 1, ?, 3, 'c', '', 'hidden', 'bad', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`, member.ID); err != nil {
		t.Fatal(err)
	}
	auth := login(t, app, "member-appeal", "MemberPass123!")
	req := httptest.NewRequest(http.MethodPost, "/api/reviews/50/appeal", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	addAuth(req, auth)
	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200 appeal, got %d body=%s", resp.StatusCode, string(body))
	}
	var status string
	if err := st.DB.QueryRow(`SELECT appeal_status FROM reviews WHERE id = 50`).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != "pending" {
		t.Fatalf("expected appeal_status=pending, got %s", status)
	}
}

func TestAppealReviewRejectedForNonAuthor(t *testing.T) {
	app, st := setupApp(t)
	defer st.Close()
	authSvc := services.NewAuthService(st, 30*time.Minute, 5, 15*time.Minute)
	authorHash, _ := authSvc.HashPassword("AuthorPass123!")
	if err := st.CreateUser("review-author", authorHash, "member", nil); err != nil {
		t.Fatal(err)
	}
	otherHash, _ := authSvc.HashPassword("OtherPass1234!")
	if err := st.CreateUser("other-member", otherHash, "member", nil); err != nil {
		t.Fatal(err)
	}
	author, _ := st.FindUserByUsername("review-author")
	if _, err := st.DB.Exec(`INSERT INTO reviews (id, club_id, fulfilled_order_id, site_id, member_id, reviewer_id, stars, comment, tags, appeal_status, hidden_reason, hidden_at, created_at) VALUES (51, 1, NULL, 1, 1, ?, 3, 'c', '', 'hidden', 'bad', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`, author.ID); err != nil {
		t.Fatal(err)
	}
	auth := login(t, app, "other-member", "OtherPass1234!")
	req := httptest.NewRequest(http.MethodPost, "/api/reviews/51/appeal", nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	addAuth(req, auth)
	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode == 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected non-200 for non-author appeal, got %d body=%s", resp.StatusCode, string(body))
	}
}

// --- POST /api/mdm/dimensions/import ---

func TestImportDimensionCSV(t *testing.T) {
	app, st := setupApp(t)
	defer st.Close()
	adminHash, _ := services.NewAuthService(st, 30*time.Minute, 5, 15*time.Minute).HashPassword("StrongAdmin123!")
	if err := st.UpdatePassword(1, adminHash, false); err != nil {
		t.Fatal(err)
	}
	auth := login(t, app, "admin", "StrongAdmin123!")
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("dimension_name", "region")
	_ = writer.WriteField("version_label", "rg-v1")
	part, err := writer.CreateFormFile("file", "region.csv")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write([]byte("code,name\nUSCA,California\n")); err != nil {
		t.Fatal(err)
	}
	_ = writer.Close()
	req := httptest.NewRequest(http.MethodPost, "/api/mdm/dimensions/import", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	addAuth(req, auth)
	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200 import, got %d body=%s", resp.StatusCode, string(b))
	}
}

func TestImportDimensionMissingFile(t *testing.T) {
	app, st := setupApp(t)
	defer st.Close()
	adminHash, _ := services.NewAuthService(st, 30*time.Minute, 5, 15*time.Minute).HashPassword("StrongAdmin123!")
	if err := st.UpdatePassword(1, adminHash, false); err != nil {
		t.Fatal(err)
	}
	auth := login(t, app, "admin", "StrongAdmin123!")
	req := httptest.NewRequest(http.MethodPost, "/api/mdm/dimensions/import", nil)
	addAuth(req, auth)
	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 400 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 400 missing file, got %d body=%s", resp.StatusCode, string(body))
	}
}

// --- POST /api/mdm/sales-facts/import ---

func TestImportSalesFactsRejectsUnknownDimensionCodes(t *testing.T) {
	app, st := setupApp(t)
	defer st.Close()
	adminHash, _ := services.NewAuthService(st, 30*time.Minute, 5, 15*time.Minute).HashPassword("StrongAdmin123!")
	if err := st.UpdatePassword(1, adminHash, false); err != nil {
		t.Fatal(err)
	}
	auth := login(t, app, "admin", "StrongAdmin123!")
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", "sales.csv")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write([]byte("product_code,customer_code,channel_code,region_code,time_code,amount,transaction_date\nPR001,CU001,CH001,RG01,TM001,100.0,2026-03-01\n")); err != nil {
		t.Fatal(err)
	}
	_ = writer.Close()
	req := httptest.NewRequest(http.MethodPost, "/api/mdm/sales-facts/import", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	addAuth(req, auth)
	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatal(err)
	}
	// No dimensions seeded, so unknown codes should fail validation
	if resp.StatusCode == 200 {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected non-200 when dimension codes unknown, got %d body=%s", resp.StatusCode, string(b))
	}
}

func TestImportSalesFactsMissingFile(t *testing.T) {
	app, st := setupApp(t)
	defer st.Close()
	adminHash, _ := services.NewAuthService(st, 30*time.Minute, 5, 15*time.Minute).HashPassword("StrongAdmin123!")
	if err := st.UpdatePassword(1, adminHash, false); err != nil {
		t.Fatal(err)
	}
	auth := login(t, app, "admin", "StrongAdmin123!")
	req := httptest.NewRequest(http.MethodPost, "/api/mdm/sales-facts/import", nil)
	addAuth(req, auth)
	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 400 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 400 missing file, got %d body=%s", resp.StatusCode, string(body))
	}
}

// --- POST /api/members/:id ---

func TestUpdateMemberByOrganizer(t *testing.T) {
	app, st := setupApp(t)
	defer st.Close()
	authSvc := services.NewAuthService(st, 30*time.Minute, 5, 15*time.Minute)
	hash, _ := authSvc.HashPassword("OrganizerPass123!")
	if err := st.CreateUser("org-update-mem", hash, "organizer", int64Ptr(1)); err != nil {
		t.Fatal(err)
	}
	memberID, err := st.InsertMember(storeMember(1, "enc", "enc", "Old Name"))
	if err != nil {
		t.Fatal(err)
	}
	auth := login(t, app, "org-update-mem", "OrganizerPass123!")
	form := url.Values{}
	form.Set("full_name", "New Name")
	form.Set("email", "new@example.com")
	form.Set("phone", "555-0000")
	form.Set("join_date", "2026-03-01")
	form.Set("position_title", "Lead")
	form.Set("is_active", "true")
	form.Set("group_name", "Beta")
	form.Set("custom_fields", "{}")
	req := httptest.NewRequest(http.MethodPost, "/api/members/"+strconv.FormatInt(memberID, 10), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	addAuth(req, auth)
	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200 update, got %d body=%s", resp.StatusCode, string(body))
	}
	var fullName string
	if err := st.DB.QueryRow(`SELECT full_name FROM members WHERE id = ?`, memberID).Scan(&fullName); err != nil {
		t.Fatal(err)
	}
	if fullName != "New Name" {
		t.Fatalf("expected full_name=New Name, got %s", fullName)
	}
}

func TestUpdateMemberRejectedForOtherClub(t *testing.T) {
	app, st := setupApp(t)
	defer st.Close()
	if _, err := st.DB.Exec(`INSERT OR IGNORE INTO clubs (id, name) VALUES (2, 'Club Two')`); err != nil {
		t.Fatal(err)
	}
	authSvc := services.NewAuthService(st, 30*time.Minute, 5, 15*time.Minute)
	hash, _ := authSvc.HashPassword("OrganizerPass123!")
	if err := st.CreateUser("org-club1", hash, "organizer", int64Ptr(1)); err != nil {
		t.Fatal(err)
	}
	memberID, err := st.InsertMember(storeMember(2, "enc", "enc", "Target"))
	if err != nil {
		t.Fatal(err)
	}
	auth := login(t, app, "org-club1", "OrganizerPass123!")
	form := url.Values{}
	form.Set("full_name", "Attacker")
	form.Set("email", "x@example.com")
	form.Set("phone", "555")
	form.Set("join_date", "2026-03-01")
	req := httptest.NewRequest(http.MethodPost, "/api/members/"+strconv.FormatInt(memberID, 10), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	addAuth(req, auth)
	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 403 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 403 cross-club update, got %d body=%s", resp.StatusCode, string(body))
	}
}

// --- GET /api/flags/evaluate/:key ---

func TestFlagEvaluateAdminAccess(t *testing.T) {
	app, st := setupApp(t)
	defer st.Close()
	adminHash, _ := services.NewAuthService(st, 30*time.Minute, 5, 15*time.Minute).HashPassword("StrongAdmin123!")
	if err := st.UpdatePassword(1, adminHash, false); err != nil {
		t.Fatal(err)
	}
	auth := login(t, app, "admin", "StrongAdmin123!")
	req := httptest.NewRequest(http.MethodGet, "/api/flags/evaluate/credit_engine_v2", nil)
	addAuth(req, auth)
	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200 flag evaluate, got %d body=%s", resp.StatusCode, string(body))
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "\"flag\"") {
		t.Fatalf("expected flag key in response, got %s", string(body))
	}
}

func TestFlagEvaluateRejectsNonAdmin(t *testing.T) {
	app, st := setupApp(t)
	defer st.Close()
	authSvc := services.NewAuthService(st, 30*time.Minute, 5, 15*time.Minute)
	hash, _ := authSvc.HashPassword("MemberPass123!")
	if err := st.CreateUser("member-flag", hash, "member", nil); err != nil {
		t.Fatal(err)
	}
	auth := login(t, app, "member-flag", "MemberPass123!")
	req := httptest.NewRequest(http.MethodGet, "/api/flags/evaluate/credit_engine_v2", nil)
	addAuth(req, auth)
	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 403 {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 403 non-admin flag evaluate, got %d body=%s", resp.StatusCode, string(body))
	}
}
