<div align="center">

# gufu
Uma biblioteca em Go para interagir com a API da UFU.

<!-- badges -->
<a href="https://pkg.go.dev/github.com/data-ru/gufu" title="Go API Reference" rel="nofollow"><img src="https://img.shields.io/badge/go-documentação-blue.svg?style=for-the-badge" alt="Go API Reference"></a>

</div>

## Funcionalidades

### LoginViaSSO(usuário, senha string)
Realiza o login no sistema da UFU usando a API do SSO. Retorna um ponteiro para `DadosSSO` e um erro.

### ObterIdUfu(id string)
Obtém as informações de uma identidade digital da UFU. Retorna um ponteiro para `IdUfu` e um erro.

### ObterTodosOsCardapios()
Obtém todos os cardápios de refeições da UFU. Retorna um slice de `Cardapio` e um erro.

### ObterCardapiosFuturosPorCampus(campus string)
Obtém os cardápios futuros de refeições da UFU para um campus específico. Retorna um slice de `Cardapio` e um erro.

### ObterCardapioPorCampus(campus string)
Obtém o cardápio de refeições da UFU para um campus específico. Retorna um ponteiro para `Cardapio` e um erro.

### Descriptografar(texto)
Descriptografa as respostas da API do aplicativo móvel da UFU. Retorna um JSON descriptografado e um erro.

### Criptografar(texto)
Prepara um JSON para ser enviado para a API do aplicativo móvel da UFU. Retorna um JSON criptografado e um erro.
