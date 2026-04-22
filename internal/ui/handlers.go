package ui

import (
	"log/slog"
	"net"
	"net/http"
	"time"

	"gitea.homelab.local/facilpass/vouchers-app/internal/auth"
)

type UIOptions struct {
	AdminUser       string
	AdminPassBcrypt string
	Sessions        *auth.SessionManager
	RateLimiter     *auth.LoginRateLimiter
	CSRF            *CSRFManager
	Templates       Renderer
	Logger          *slog.Logger
	CookieName      string
	CookieSecure    bool
}

type Renderer interface {
	Render(w http.ResponseWriter, name string, data map[string]any) error
}

type UI struct {
	o UIOptions
}

func New(o UIOptions) *UI {
	if o.CookieName == "" {
		o.CookieName = "vouchers_session"
	}
	if o.Logger == nil {
		o.Logger = slog.Default()
	}
	return &UI{o: o}
}

func (u *UI) GetLogin(w http.ResponseWriter, r *http.Request) {
	_ = u.o.Templates.Render(w, "login", map[string]any{
		"Title": "Entrar",
		"Error": r.URL.Query().Get("err"),
	})
}

func (u *UI) PostLogin(w http.ResponseWriter, r *http.Request) {
	ip := clientIP(r)
	if !u.o.RateLimiter.Allow(ip) {
		http.Redirect(w, r, "/admin?err=bloqueado", http.StatusSeeOther)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "form inválida", http.StatusBadRequest)
		return
	}
	user := r.FormValue("user")
	pass := r.FormValue("pass")
	if user != u.o.AdminUser || !auth.VerifyPassword(u.o.AdminPassBcrypt, pass) {
		u.o.Logger.Warn("login failed", slog.String("ip", ip), slog.String("user", user))
		http.Redirect(w, r, "/admin?err=credenciais", http.StatusSeeOther)
		return
	}
	token, err := u.o.Sessions.Sign(user)
	if err != nil {
		http.Error(w, "session sign", http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     u.o.CookieName,
		Value:    token,
		Path:     "/admin",
		HttpOnly: true,
		Secure:   u.o.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(8 * time.Hour),
	})
	u.o.Logger.Info("login ok", slog.String("ip", ip), slog.String("user", user))
	http.Redirect(w, r, "/admin/app", http.StatusSeeOther)
}

func (u *UI) GetApp(w http.ResponseWriter, r *http.Request) {
	user, err := u.sessionUser(r)
	if err != nil {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}
	csrf := u.o.CSRF.Generate(user)
	_ = u.o.Templates.Render(w, "app", map[string]any{
		"Title": "Upload",
		"CSRF":  csrf,
		"User":  user,
	})
}

func (u *UI) GetLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:   u.o.CookieName,
		Value:  "",
		Path:   "/admin",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (u *UI) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := u.sessionUser(r); err != nil {
			http.Redirect(w, r, "/admin", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (u *UI) CSRFCheckFromRequest(r *http.Request) func(string) bool {
	user, err := u.sessionUser(r)
	if err != nil {
		return func(string) bool { return false }
	}
	return func(token string) bool {
		return u.o.CSRF.Verify(user, token)
	}
}

func (u *UI) sessionUser(r *http.Request) (string, error) {
	c, err := r.Cookie(u.o.CookieName)
	if err != nil {
		return "", err
	}
	return u.o.Sessions.Verify(c.Value)
}

func clientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		if comma := len(ip); comma > 0 {
			for i, c := range ip {
				if c == ',' {
					return ip[:i]
				}
			}
		}
		return ip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
