package ufutil

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"strings"
	"time"
)

var (
	ClienteHTTP = http.Client{
		Timeout: 15 * time.Second,
	} //Cliente HTTP usado em todas as requsições, com timeout de 15 segundos. Pode ser alterado para atender necessidades específicas.
	userAgent    = fmt.Sprintf("ufutil/v2.0.0 +(https://github.com/data-ru/ufutil; go/%v; %v/%v)", runtime.Version(), runtime.GOOS, runtime.GOARCH) //User-Agent padrão usado em todas as requisições. Se parece algo como: ufutil/v2.0.0 +(https://github.com/data-ru/ufutil; go/1.23.4; windows/amd64);
	ssoUrl       = "https://sso.ufu.br"                                                                                                             //URL Base do SSO da UFU. Não deve ser alterado.
	mobileApiUrl = "https://www.sistemas.ufu.br/mobile-gateway"                                                                                     //URL Base da API do aplicativo móvel da UFU.
)

// DadosSSO é a estrutura que contém as informações do usuário autenticado no SSO. É retornado na função LoginViaSSO.
type DadosSSO struct {
	Cpf           string         `json:"cpf"`             //CPF do usuário
	Nome          string         `json:"nome"`            //Nome do usuário
	Chave         string         `json:"chave"`           //Token de autenticação
	Email         string         `json:"email"`           //Email do usuário
	ExpiraEm      int            `json:"expira_em"`       //Data de expiração do token de autenticação, Unix timestamp em milisegundos
	IDPessoa      int            `json:"id_pessoa"`       //ID interno da pessoa
	EmitidoEm     int            `json:"emitido_em"`      //Data de emissão do token de autenticação , Unix timestamp em milisegundos
	AccessTokenID string         `json:"access_token_id"` //ID do token de autenticação
	Roles         []string       `json:"roles"`           //Perfis do usuário
	Perfis        string         `json:"perfis"`          //Quantidade de perfis do usuário
	Cookies       []*http.Cookie //Cookies da sessão
}

// LoginViaSSO é a função que realiza o login no sistema da UFU usando a api do SSO. Retorna um ponteiro para DadosSSO e um erro.
func LoginViaSSO(email, senha string) (*DadosSSO, error) {
	if email == "" || senha == "" {
		return nil, errors.New("usuario ou senha estão vazios")
	}
	dadosDoLogin, err := json.Marshal(map[string]string{
		"uid":   email,
		"senha": senha,
	})
	if err != nil {
		return nil, err
	}

	//Primeira parte do login:
	//Envia um POST para /autenticar com o usuário e senha
	requestCreateLogin, err := http.NewRequest(http.MethodPost, ssoUrl+"/autenticar", bytes.NewReader(dadosDoLogin))
	if err != nil {
		return nil, err
	}
	requestCreateLogin.Header.Add("Content-Type", "application/json")
	requestCreateLogin.Header.Add("User-Agent", userAgent)

	responseCreateLogin, err := ClienteHTTP.Do(requestCreateLogin)
	if err != nil {
		return nil, err
	}
	defer responseCreateLogin.Body.Close()

	body, err := io.ReadAll(responseCreateLogin.Body)
	if err != nil {
		return nil, err
	}
	//O servidor nos retorna o seguinte conteúdo:
	// /cliente-login?t=XXXXXXXXXXXX
	responseCreateLoginBody := string(body)

	//Se for bem sucedido, o servidor retorna o status 201 com um JSON contendo a URL para obter as informações do usuário
	if responseCreateLogin.StatusCode != 201 {
		return nil, fmt.Errorf("algo deu errado, status http: %v, mensagem do servidor: %v", responseCreateLogin.StatusCode, responseCreateLoginBody)
	}

	//Segunda parte do login:
	//Então substituimos /cliente-login?t=XXXXXXXX por /usuario?t=XXXXXXXXXX
	// A página /cliente-login mostra os serviços acessiveis com o id ufu,
	// já /usuario?t=XXXXXXXXXX mostra as informações do usuário em um json.
	getUserPath := strings.ReplaceAll(responseCreateLoginBody, "cliente-login", "usuario")

	//E enviamos uma requisição GET para /usuario?t=XXXXXXXXXX com os cookies da requisição anterior
	requestGetUser, err := http.NewRequest(http.MethodGet, ssoUrl+getUserPath, nil)
	if err != nil {
		return nil, err
	}
	cookiesCreate := responseCreateLogin.Cookies()
	for _, v := range cookiesCreate {
		requestGetUser.AddCookie(v) //Adiciona os cookies da request anterior
	}
	requestGetUser.Header.Add("User-Agent", userAgent)

	responseGetUser, err := ClienteHTTP.Do(requestGetUser)
	if err != nil {
		return nil, err
	}
	defer responseGetUser.Body.Close()

	body, err = io.ReadAll(responseGetUser.Body)
	if err != nil {
		return nil, err
	}

	if responseGetUser.StatusCode != 200 {
		return nil, fmt.Errorf("algo deu errado ao obter as informações do usuario, status http %v, (%v)", responseGetUser.Status, err)
	}

	var informaçõesUsuario DadosSSO
	err = json.Unmarshal(body, &informaçõesUsuario)
	if err != nil {
		return nil, err
	}

	//Adiciona os cookies da requisição anterior
	informaçõesUsuario.Cookies = responseCreateLogin.Cookies()

	return &informaçõesUsuario, nil
}

// A estrutura IdUfu contém as informações de uma identidade digital da UFU. É retornado na função ObterIdUfu.
type IdUfu struct {
	ID                int    `json:"id"`                //ID da identidade digital
	Matricula         string `json:"matricula"`         //Matrícula do aluno
	Nome              string `json:"nome"`              //Nome do aluno
	NomePai           string `json:"nomePai"`           //Nome do pai do aluno. Pode ser vazio
	NomeMae           string `json:"nomeMae"`           //Nome da mãe do aluno.
	RG                string `json:"rg"`                //RG do aluno
	OrgaoEmissor      string `json:"orgaoEmissor"`      //Órgão emissor do RG
	CPF               string `json:"cpf"`               //CPF do aluno
	Naturalidade      string `json:"naturalidade"`      //Naturalidade do aluno
	Vinculo           string `json:"vinculo"`           //Vínculo (aluno, servidor, etc)
	DataNascimento    int64  `json:"dataNascimento"`    //Data de nascimento do aluno, Unix timestamp em milisegundos
	CodigoBarra       string `json:"codigoBarra"`       //Dado do QR Code da identidade digital
	Informacao        string `json:"informacao"`        //Curso, tipo de curso e turno (Ex: "Graduação em Sistemas de Informação: Bacharelado - Noturno")
	SituacaoDescricao string `json:"situacaoDescricao"` //Situação da identidade digital
	Situacao          int    `json:"situacao"`          //Situação da identidade digital, 3 = Ativa
	Foto              string `json:"foto"`              //Foto em base64
}

// Estrutura intermediária usada para verificar se a identidade digital não é nula.
type resultadoIdentidade struct {
	IdUfu                    *IdUfu  `json:"identidadeDigital"`        //Chave com a identidade digital
	DocumentoArquivoTOResult any     `json:"documentoArquivoTOResult"` // Campo desconhecido
	DataNascimentoString     *string `json:"dataNascimentoString"`     //Data de nascimento do aluno em string (Ex: "01/01/2000")
}

// ObterIdUfu é a função que obtém as informações de uma identidade digital da UFU. Retorna um ponteiro para IdUfu e um erro.
// O parâmetro id é o número da identidade digital, presente no QR Code. Por exemplo, para o QR Code "https://www.sistemas.ufu.br/valida-ufu/#/id-digital/123123456789", o id é "123123456789".
func ObterIdUfu(id string) (*IdUfu, error) {
	//Envia uma requisição GET para /buscarDadosIdDigital?idIdentidade=ID
	res, err := requisiçãoGenerica("https://www.sistemas.ufu.br/valida-gateway/id-digital/buscarDadosIdDigital?idIdentidade="+id, http.MethodGet, nil)
	if err != nil {
		return nil, err
	}
	corpo, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("algo deu errado ao verificar a identidade, status http: %v", res.StatusCode)
	}

	var jsonId resultadoIdentidade
	err = json.Unmarshal(corpo, &jsonId)
	if err != nil {
		return nil, err
	}

	//Se o número da identidade digital não for encontrado, o servidor retorna NULL nos 3 campos.
	if jsonId.DataNascimentoString == nil {
		return nil, errors.New("identidade invalida")
	}

	return jsonId.IdUfu, nil
}

// A estrutura Cardapio contém as informações de um cardápio de refeições da UFU. É retornado na função ObterTodosOsCardapios.
type Cardapio struct {
	Titulo            string `json:"titulo"`             //Título do restaurante (Ex: 2024/12/16 - Cardápio Restaurante Universitário - Santa Mônica)
	Local             string `json:"local"`              //Local do restaurante (Ex: Restaurante Universitário - Santa Mônica)
	Mensagem          string `json:"mensagem"`           //Alguma mensagem do restaurante, não parece ser usado
	PrincipalAlmoco   string `json:"principal_almoco"`   //Proteina principal do almoço
	VegetarianoAlmoco string `json:"vegetariano_almoco"` //Proteina vegetariana do almoço
	ArrozAlmoco       string `json:"arroz_almoco"`       //Tipo de arroz do almoço
	FeijaoAlmoco      string `json:"feijao_almoco"`      //Tipo de feijão do almoço
	GuarnicaoAlmoco   string `json:"guarnicao_almoco"`   //Guarnição do almoço
	SaladaAlmoco      string `json:"salada_almoco"`      //Salada do almoço
	SobremesaAlmoco   string `json:"sobremesa_almoco"`   //Sobremesa do almoço
	SucoAlmoco        string `json:"suco_almoco"`        //Suco do almoço
	Data              string `json:"data"`               //Data do cardápio (Ex: 16/12/2024)
	PrincipalJantar   string `json:"principal_jantar"`   //Proteina principal do jantar
	VegetarianoJantar string `json:"vegetariano_jantar"` //Proteina vegetariana do jantar
	ArrozJantar       string `json:"arroz_jantar"`       //Tipo de arroz do jantar
	FeijaoJantar      string `json:"feijao_jantar"`      //Tipo de feijão do jantar
	GuarnicaoJantar   string `json:"guarnicao_jantar"`   //Guarnição do jantar
	SaladaJantar      string `json:"salada_jantar"`      //Salada do jantar
	SobremesaJantar   string `json:"sobremesa_jantar"`   //Sobremesa do jantar
	SucoJantar        string `json:"suco_jantar"`        //Suco do jantar
	Nid               string `json:"nid"`                //ID interno do cardápio
}

var ErrNãoHáRefeições = errors.New("não há refeições agendadas para hoje") //Erro retornado quando não há refeições agendadas para hoje.

// ObterTodosOsCardapios é a função que obtém todos os cardápios de refeições da UFU. Retorna um slice de Cardapio e um erro.
func ObterTodosOsCardapios() ([]Cardapio, error) {
	resp, err := requisiçãoGenerica(mobileApiUrl+"/api/cardapios/", http.MethodGet, nil)
	if err != nil {
		return nil, err
	}
	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	decryptBody, err := Descriptografar(string(bodyResp))
	if err != nil {
		return nil, err
	}

	if decryptBody == "{}" {
		return nil, ErrNãoHáRefeições
	}

	var cardapios []Cardapio
	err = json.Unmarshal([]byte(decryptBody), &cardapios)
	if err != nil {
		return nil, err
	}

	return cardapios, nil
}

type Campus struct {
	ID   int
	Nome string
}

var Campi = map[string]Campus{
	"sm": {ID: 277, Nome: "Santa Mônica"},
	"um": {ID: 279, Nome: "Umuarama"},
	"gl": {ID: 1747, Nome: "Glória"},
	"po": {ID: 278, Nome: "Pontal"},
	"mc": {ID: 6097, Nome: "Monte Carmelo"},
}

func ObterCardapiosFuturosPorCampus(campus string) ([]Cardapio, error) {
	campusID, ok := Campi[campus]
	if !ok {
		return nil, errors.New("campus inválido")
	}

	resp, err := requisiçãoGenerica(fmt.Sprintf("%s/api/proximos-cardapios/%d", mobileApiUrl, campusID.ID), http.MethodGet, nil)
	if err != nil {
		return nil, err
	}
	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	decryptBody, err := Descriptografar(string(bodyResp))
	if err != nil {
		return nil, err
	}

	if decryptBody == "{}" {
		return nil, ErrNãoHáRefeições
	}

	var cardapios []Cardapio
	err = json.Unmarshal([]byte(decryptBody), &cardapios)
	if err != nil {
		return nil, fmt.Errorf("%v: %v", err, decryptBody)
	}

	return cardapios, nil
}

func ObterCardapioPorCampus(campus string) (*Cardapio, error) {
	campusID, ok := Campi[campus]
	if !ok {
		return nil, errors.New("campus inválido")
	}

	resp, err := requisiçãoGenerica(fmt.Sprintf("%s/api/cardapios/%d", mobileApiUrl, campusID.ID), http.MethodGet, nil)
	if err != nil {
		return nil, err
	}
	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	decryptBody, err := Descriptografar(string(bodyResp))
	if err != nil {
		return nil, err
	}

	if decryptBody == "{}" {
		return nil, ErrNãoHáRefeições
	}

	var cardapio Cardapio
	err = json.Unmarshal([]byte(decryptBody), &cardapio)
	if err != nil {
		return nil, err
	}

	return &cardapio, nil
}

// A função Descriptografar descriptografa a respostas da API do aplicativo móvel da UFU. Retorna um JSON descriptografado e um erro.
func Descriptografar(corpo string) (string, error) {
	return decryptEncodedJson(corpo)
}

// A função Criptografar prepara um JSON (requestParams) para ser enviado para a API do aplicativo móvel da UFU. Retorna um JSON criptografado e um erro.
func Criptografar(json string) (string, error) {
	return criptografarRequestParams(json)
}

// Função genérica para fazer requisições HTTP. Não pode ser usada diretamente.
func requisiçãoGenerica(url, meteodo string, corpo io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(meteodo, url, corpo)
	if err != nil {
		return nil, err
	}
	//req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", userAgent)

	res, err := ClienteHTTP.Do(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}