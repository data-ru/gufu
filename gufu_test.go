package gufu

import (
	"encoding/json"
	"os"
	"testing"

	_ "github.com/joho/godotenv/autoload"
)

func TestLogin(t *testing.T) {
	user, pass := os.Getenv("UFU_USER"), os.Getenv("UFU_PASS")
	a, err := LoginViaSSO(user, pass)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(a)
}

func TestLoginMobile(t *testing.T) {
	u, p := os.Getenv("UFU_USER"), os.Getenv("UFU_PASS")
	logindados, err := LoginViaMobile(u, p)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(logindados)
}

func TestObterCarteirinhaMobile(t *testing.T) {
	u, p := os.Getenv("UFU_USER"), os.Getenv("UFU_PASS")
	login, err := LoginViaMobile(u, p)
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Login ok...")
	c, err := login.BuscarIdentidadeDigital()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(c.Matricula)
}

func TestCardapioCampi(t *testing.T) {
	a, err := ObterTodosOsCardapios()
	if err != nil {
		if err == ErrNãoHáRefeições {
			t.Log("não há refeições")
			t.SkipNow()
		}
		t.Fatal(err)
	}
	t.Log(a)
}

func TestCardapioCampiSantaMonica(t *testing.T) {
	a, err := ObterCardapioPorCampus("sm")
	if err != nil {
		if err == ErrNãoHáRefeições {
			t.Log("não há refeições")
			t.SkipNow()
		}
		t.Fatal(err)
	}
	t.Log(a)
}

func TestCardapioFuturoCampiSantaMonica(t *testing.T) {
	a, err := ObterCardapiosFuturosPorCampus("sm")
	if err != nil {
		if err == ErrNãoHáRefeições {
			t.Log("não há refeições")
			t.SkipNow()
		}
		t.Fatal(err)
	}
	t.Logf("Almoço de amanhã (%v): %v de proteina, guarnição é: %v", a[1].Data, a[1].PrincipalAlmoco, a[1].GuarnicaoAlmoco)
}

func TestValidarId(t *testing.T) {
	v, err := ObterIdUfu(os.Getenv("UFU_ID"))
	if err != nil {
		t.Fatal(err)
	}
	e, _ := json.Marshal(v)
	t.Logf("%s", e)
}

func TestDescriptografar(t *testing.T) {
	tt := `qb/ItbJgpIk88xrKTjGpKQ==G2b1UFYMYjNViPZY6bSpvHnNYxHHjVSaWQaJcBxegnjiODm3vs2a`
	v, err := Descriptografar(tt)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s", v)

}

func TestCriptografar(t *testing.T) {
	tc := `test`
	v, err := Criptografar(tc)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s", v)
}
