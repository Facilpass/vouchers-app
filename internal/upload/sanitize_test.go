package upload

import "testing"

func TestSanitizeFilename(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"VOUCHER_BARRA.jpeg", "VOUCHER_BARRA.jpeg"},
		{"Voucher Barra.jpeg", "Voucher_Barra.jpeg"},
		{"Açaí da Copa (2026)!.jpeg", "Acai_da_Copa_2026.jpeg"},
		{"'; DROP TABLE users--.jpg", "DROPTABLEusers--.jpg"},
		{"<script>alert(1)</script>.jpg", "scriptalert1script.jpg"},
		{"file\x00.jpg", "file.jpg"},
		{"..foo.jpg", "foo.jpg"},
		{".htaccess", "htaccess"},
		{"foo___bar---baz.jpg", "foo_bar-baz.jpg"},
		{"Ingresso Seção Niterói.png", "Ingresso_Secao_Niteroi.png"},
	}
	for _, tc := range cases {
		got := SanitizeFilename(tc.in)
		if got != tc.want {
			t.Errorf("SanitizeFilename(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestSanitizePath(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"loungebrahma/copadobrasil/camarote", "loungebrahma/copadobrasil/camarote"},
		{"LoungeBrahma/Copa do Brasil/Camarote VIP", "loungebrahma/copa-do-brasil/camarote-vip"},
		{"/leading/slash", "leading/slash"},
		{"trailing/slash/", "trailing/slash"},
		{"double//slash", "double/slash"},
		{"../../etc/passwd", "etc/passwd"},
		{"loungebrahma/ação", "loungebrahma/acao"},
	}
	for _, tc := range cases {
		got := SanitizePath(tc.in)
		if got != tc.want {
			t.Errorf("SanitizePath(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestValidPathRegex(t *testing.T) {
	valid := []string{
		"tenant/evento/categoria",
		"a/b",
		"single",
		"with-hyphen/and-123",
	}
	for _, p := range valid {
		if !IsValidPath(p) {
			t.Errorf("IsValidPath(%q) = false, want true", p)
		}
	}
	invalid := []string{
		"UPPER/case",
		"espaço aqui",
		"ponto.aqui",
		"/leading",
		"trailing/",
		"",
		"a/b/c/d/e/f/g/h/i/j/k",
	}
	for _, p := range invalid {
		if IsValidPath(p) {
			t.Errorf("IsValidPath(%q) = true, want false", p)
		}
	}
}

func TestValidExt(t *testing.T) {
	valid := []string{".webp", ".jpeg", ".jpg", ".png"}
	for _, e := range valid {
		if !IsValidExt(e) {
			t.Errorf("IsValidExt(%q) = false, want true", e)
		}
	}
	invalid := []string{".gif", ".bmp", ".pdf", ".svg", ".php", ""}
	for _, e := range invalid {
		if IsValidExt(e) {
			t.Errorf("IsValidExt(%q) = true, want false", e)
		}
	}
}
