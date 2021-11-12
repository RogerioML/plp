package plp

import (
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/encoding/charmap"
)

const (
	nuCliente = 2019020716
	//Versao indica a versão atual do módulo de manipulação de PLPs
	Versao      = "1.1.12"
	LayoutMysql = "2006-01-02 15:04:05"
)

/*
ErrPlpNaoVazia variável com mensagem de erro módulo PLP
*/
var (
	ErrPlpNaoVazia          = errors.New("negocio: ID da PLP deve estar vazio")
	erCartao                *regexp.Regexp
	ErrCartaoInvalido       = errors.New("negocio: formato de cartão inválido")
	ErrContratoInvalido     = errors.New("negocio: formato de contrato inválido")
	erDr                    *regexp.Regexp
	ErrDrInvalido           = errors.New("negocio: formato de diretoria inválido")
	erCodAdm                *regexp.Regexp
	ErrCodAdmInvalido       = errors.New("negocio: formato de código administrativo inválido")
	ErrNomeRemetente        = errors.New("negocio: formato de nome de remetente inválido")
	ErrLogradouroRemetente  = errors.New("negocio: formato de logradouro de remetente inválido")
	ErrNumeroRemetente      = errors.New("negocio: formato de número de remetente inválido")
	ErrComplementoRemetente = errors.New("negocio: formato de complemento de remetente inválido")
	ErrBairroRemetente      = errors.New("negocio: formato de bairro de remetente inválido")
	erCep                   *regexp.Regexp
	ErrCepRemetente         = errors.New("negocio: formato de CEP de remetente inválido")
	ErrCidadeRemetente      = errors.New("negocio: formato de cidade de remetente inválido")
	erUf                    *regexp.Regexp
	ErrUfRemetente          = errors.New("negocio: formato de UF de remetente inválido")
	erTelefone              *regexp.Regexp
	ErrTelefoneRemetente    = errors.New("negocio: formato de telefone de remetente inválido")
	erEmail                 *regexp.Regexp
	ErrEmailRemetente       = errors.New("negocio: formato de email do remetente inválido")
	ErrCelularRemetente     = errors.New("negocio: formato de celular de remetente inválido")
	Wsdl                    string
	User                    string
	Pass                    string
	RegexEtiqueta           *regexp.Regexp
)

func init() {
	erCartao = regexp.MustCompile(`^[0-9]{10}$`)
	erDr = regexp.MustCompile(`^[0-9]{2}$`)
	erCodAdm = regexp.MustCompile(`^[0-9]{8}$`)
	erCep = regexp.MustCompile(`^[0-9]{8}$`)
	erUf = regexp.MustCompile(`^[a-zA-Z]{2}$`)
	erTelefone = regexp.MustCompile(`^[0-9]*$`)
	erEmail = regexp.MustCompile(`^[a-zA-Z0-9.!#$%&’*+/=?^_{|}~-]+@[a-zA-Z0-9-]+(?:\.[a-zA-Z0-9-]+)*$`)
	RegexEtiqueta = regexp.MustCompile(`^[A-Z]{2}[1-9]{9}[A-Z]{2}$`)
}

// Plp estrutura da PLP
type Plp struct {
	ID                 int       `xml:"-"`
	Status             int       `xml:"-"`
	NumeroCliente      int       `xml:"-"`
	Postagem           time.Time `xml:"-"`
	PostagemSara       time.Time `xml:"-"`
	Fechamento         time.Time `xml:"-"`
	AtualizacaoCliente time.Time `xml:"-"`
	XMLName            xml.Name  `xml:"correioslog"`
	TipoArquivo        string    `xml:"tipo_arquivo"`
	VersaoArquivo      string    `xml:"versao_arquivo"`
	Plp                struct {
		IDPlp              int     `xml:"id_plp" json:"id_plp"`
		ValorGlobal        float64 `xml:"valor_global"`
		McuUnidadePostagem string  `xml:"mcu_unidade_postagem"`
		NmeUnidadePostagem string  `xml:"nome_unidade_postagem"`
		CartaoPostagem     string  `xml:"cartao_postagem"`
	} `xml:"plp"`
	Remetente struct {
		NumeroContrato       string `xml:"numero_contrato"`
		NumeroDiretoria      string `xml:"numero_diretoria"`
		CodigoAdministrativo string `xml:"codigo_administrativo"`
		NomeRemetente        struct {
			CData string `xml:",cdata"`
		} `xml:"nome_remetente"`
		LogradouroRemetente struct {
			CData string `xml:",cdata"`
		} `xml:"logradouro_remetente"`
		NumeroRemetente struct {
			CData string `xml:",cdata"`
		} `xml:"numero_remetente"`
		ComplementoRemetente struct {
			CData string `xml:",cdata"`
		} `xml:"complemento_remetente"`
		BairroRemetente struct {
			CData string `xml:",cdata"`
		} `xml:"bairro_remetente"`
		CepRemetente struct {
			CData string `xml:",cdata"`
		} `xml:"cep_remetente"`
		CidadeRemetente struct {
			CData string `xml:",cdata"`
		} `xml:"cidade_remetente"`
		UfRemetente       string `xml:"uf_remetente"`
		TelefoneRemetente struct {
			CData string `xml:",cdata"`
		} `xml:"telefone_remetente"`
		FaxRemetente struct {
			CData string `xml:",cdata"`
		} `xml:"fax_remetente"`
		EmailRemetente struct {
			CData string `xml:",cdata"`
		} `xml:"email_remetente"`
		CelularRemetente struct {
			CData string `xml:",cdata"`
		} `xml:"celular_remetente"`
		CpfCnpjRemetente        string `xml:"cpf_cnpj_remetente"`
		CienciaConteudoProibido string `xml:"ciencia_conteudo_proibido"`
	} `xml:"remetente" json:"remetente"`
	FormaPagamento string    `xml:"forma_pagamento"`
	Objetos        []*Objeto `xml:"objeto_postal"`
}

// PlpRemetenteJSON estrutura para representação da PLP em formato JSON
type PlpRemetenteJSON struct {
	IDPlp          int    `json:"numero"`
	CartaoPostagem string `json:"cartao_postagem"`
	Remetente      struct {
		Contrato             string `json:"contrato"`
		Diretoria            string `json:"diretoria"`
		CodigoAdministrativo string `json:"codigo_administrativo"`
		Nome                 string `json:"nome"`
		Endereco             struct {
			Logradouro  string `json:"logradouro"`
			Numero      string `json:"numero"`
			Complemento string `json:"complemento"`
			Bairro      string `json:"bairro"`
			Cep         string `json:"cep"`
			Cidade      string `json:"cidade"`
			UF          string `json:"uf"`
			Telefone    string `json:"telefone"`
			Fax         string `json:"fax"`
			Email       string `json:"email"`
			Celular     string `json:"celular"`
		} `json:"endereco"`
	} `json:"remetente"`
}

// UnmarshalJSON converte a estrutura
func (p *Plp) UnmarshalJSON(b []byte) error {
	var wrap PlpRemetenteJSON
	if err := json.Unmarshal(b, &wrap); err != nil {
		return err
	}
	p.Plp.IDPlp = wrap.IDPlp
	p.Plp.CartaoPostagem = wrap.CartaoPostagem
	p.Remetente.NumeroContrato = wrap.Remetente.Contrato
	p.Remetente.NumeroDiretoria = wrap.Remetente.Diretoria
	p.Remetente.CodigoAdministrativo = wrap.Remetente.CodigoAdministrativo
	p.Remetente.NomeRemetente.CData = wrap.Remetente.Nome
	p.Remetente.LogradouroRemetente.CData = wrap.Remetente.Endereco.Logradouro
	p.Remetente.NumeroRemetente.CData = wrap.Remetente.Endereco.Numero
	p.Remetente.ComplementoRemetente.CData = wrap.Remetente.Endereco.Complemento
	p.Remetente.BairroRemetente.CData = wrap.Remetente.Endereco.Bairro
	p.Remetente.CepRemetente.CData = wrap.Remetente.Endereco.Cep
	p.Remetente.CidadeRemetente.CData = wrap.Remetente.Endereco.Cidade
	p.Remetente.UfRemetente = wrap.Remetente.Endereco.UF
	p.Remetente.TelefoneRemetente.CData = wrap.Remetente.Endereco.Telefone
	p.Remetente.FaxRemetente.CData = wrap.Remetente.Endereco.Fax
	p.Remetente.EmailRemetente.CData = wrap.Remetente.Endereco.Email
	p.Remetente.CelularRemetente.CData = wrap.Remetente.Endereco.Celular
	return nil
}

// XML monta o xml da PLP
func (p *Plp) XML() (string, error) {
	replacer := strings.NewReplacer(
		"<id_plp>0</id_plp>",
		"<id_plp/>",
		"<valor_global>0</valor_global>",
		"<valor_global/>",
		"<mcu_unidade_postagem></mcu_unidade_postagem>",
		"<mcu_unidade_postagem/>",
		"<nome_unidade_postagem></nome_unidade_postagem>",
		"<nome_unidade_postagem/>",
		"<codigo_objeto_cliente></codigo_objeto_cliente>",
		"<codigo_objeto_cliente/>",
		"<rt1></rt1>",
		"<rt1/>",
		"<rt2></rt2>",
		"<rt2/>",
		"<serie_nota_fiscal></serie_nota_fiscal>",
		"<serie_nota_fiscal/>",
		"<valor_nota_fiscal></valor_nota_fiscal>",
		"<valor_nota_fiscal/>",
		"<valor_nota_fiscal></valor_nota_fiscal>",
		"<valor_nota_fiscal/>",
		"<numero_comprovante_postagem></numero_comprovante_postagem>",
		"<numero_comprovante_postagem/>",
		"<forma_pagamento></forma_pagamento>",
		"<forma_pagamento/>",
		"<valor_cobrado>0</valor_cobrado>",
		"<valor_cobrado/>",
		"<complemento_remetente></complemento_remetente>",
		"<complemento_remetente><![CDATA[]]></complemento_remetente>",
		"<celular_remetente></celular_remetente>",
		"<celular_remetente><![CDATA[]]></celular_remetente>",
		"<complemento_destinatario></complemento_destinatario>",
		"<complemento_destinatario><![CDATA[]]></complemento_destinatario>",
		"<celular_destinatario></celular_destinatario>",
		"<celular_destinatario><![CDATA[]]></celular_destinatario>",
		"<codigo_usuario_postal></codigo_usuario_postal>",
		"<codigo_usuario_postal/>",
		"<centro_custo_cliente></centro_custo_cliente>",
		"<centro_custo_cliente/>",
		"<natureza_nota_fiscal></natureza_nota_fiscal>",
		"<natureza_nota_fiscal/>",
		"<descricao_objeto></descricao_objeto>",
		"<descricao_objeto><![CDATA[]]></descricao_objeto>",
		"<data_postagem_sara></data_postagem_sara>",
		"<data_postagem_sara/>",
		"<cpf_cnpj_remetente></cpf_cnpj_remetente>",
		"<cpf_cnpj_remetente/>",
		"<ciencia_conteudo_proibido></ciencia_conteudo_proibido>",
		"<ciencia_conteudo_proibido/>",
		"<restricao_anac></restricao_anac>",
		"<restricao_anac/>",
		"<valor_declarado></valor_declarado>", "",
		"<endereco_vizinho><![CDATA[]]></endereco_vizinho>", "",
		"<endereco_vizinho></endereco_vizinho>", "",
		"<cpf_cnpj_destinatario></cpf_cnpj_destinatario>",
		"<cpf_cnpj_destinatario/>",
	)
	b, err := xml.Marshal(p)
	if err != nil {
		return "", fmt.Errorf("plp xml 1: %s", err)
	}
	doc := string(b)
	doc = `<?xml version="1.0" encoding="ISO-8859-1" ?>` + doc
	doc = replacer.Replace(doc)
	b = []byte(doc)
	return string(b), nil
}

//NewPlp cria uma PLP com base nos bytes do XML.
func NewPlp(b []byte) (Plp, error) {
	var err error
	doc := string(b)
	doc = strings.Replace(doc, `<?xml version="1.0" encoding="ISO-8859-1" standalone="yes"?>`, "", 1)
	doc = strings.Replace(doc, `<?xml version="1.0" encoding="ISO-8859-1" ?>`, "", 1)
	doc = strings.Replace(doc, `<?xml version="1.0" encoding="ISO-8859-1"?>`, "", 1)
	doc = strings.Replace(doc, `<?xml version='1.0' encoding='ISO-8859-1'?>`, "", 1)
	doc = strings.Replace(doc, `<?xml version='1.0' encoding='ISO-8859-1' ?>`, "", 1)
	doc = strings.Replace(doc, `<?xml version="1.0" encoding="iso-8859-1" ?>`, "", 1)
	doc = strings.Replace(doc, `<?xml version="1.0" encoding="iso-8859-1"?>`, "", 1)
	doc = strings.Replace(doc, `<?xml version='1.0' encoding='iso-8859-1'?>`, "", 1)
	doc = strings.Replace(doc, `<?xml version='1.0' encoding='iso-8859-1' ?>`, "", 1)

	doc = strings.Replace(doc, `<?xml version="1.0" encoding="UTF-8"?>`, "", 1)
	plp := Plp{}
	err = xml.Unmarshal([]byte(doc), &plp)
	if plp.Plp.IDPlp != 0 {
		//return Plp{}, ErrPlpNaoVazia
	}
	plp.Remetente.NomeRemetente.CData = strings.TrimSpace(plp.Remetente.NomeRemetente.CData)
	plp.Remetente.LogradouroRemetente.CData = strings.TrimSpace(plp.Remetente.LogradouroRemetente.CData)
	plp.Remetente.NumeroRemetente.CData = strings.TrimSpace(plp.Remetente.NumeroRemetente.CData)
	plp.Remetente.BairroRemetente.CData = strings.TrimSpace(plp.Remetente.BairroRemetente.CData)
	plp.Remetente.CepRemetente.CData = strings.TrimSpace(plp.Remetente.CepRemetente.CData)
	plp.Remetente.CidadeRemetente.CData = strings.TrimSpace(plp.Remetente.CidadeRemetente.CData)
	plp.Remetente.TelefoneRemetente.CData = strings.TrimSpace(plp.Remetente.TelefoneRemetente.CData)
	for _, o := range plp.Objetos {
		etiqueta, err := EtiquetaDV(o.NumeroEtiqueta)
		if err != nil {
			return Plp{}, fmt.Errorf("plp newplp 1: %s", err)
		}
		o.NumeroEtiqueta = etiqueta
	}
	return plp, err
}

//IsoUtf8 converte de ISO para UTF-8
func IsoUtf8(b []byte) ([]byte, error) {
	r := charmap.ISO8859_1.NewDecoder().Reader(strings.NewReader(string(b)))
	return ioutil.ReadAll(r)
}

//estrutura para tratar requests que tiverem erro
type fault struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    struct {
		XMLName xml.Name
		Fault   struct {
			FaultCode   string `xml:"faultcode"`
			FaultString string `xml:"faultstring"`
		} `xml:"Fault"`
	} `xml:"Body"`
}

//estrutura para conter o retorno do método cancelarObjeto
type cancelarObjetoResponse struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    struct {
		CancelarObjetoResponse struct {
			Retorno bool `xml:"return"`
		} `xml:"cancelarObjetoResponse"`
	} `xml:"Body"`
}

//estrutura para conter o retorno do método geraDigitoVeiricadorEtiquetas
type solicitaEtiquetasResponse struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    struct {
		XMLName                   xml.Name
		SolicitaEtiquetasResponse struct {
			FaixaEtiquetas string `xml:"return"`
		} `xml:"solicitaEtiquetasResponse"`
	} `xml:"Body"`
}

//CancelarObjeto faz a chamada ao SIGEPWEB para cancelar uma etiqueta obtida anteriormente
func CancelarObjeto(etiqueta string, plp string, user string, senha string) error {
	payload := fmt.Sprintf(`
			<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:cli="http://cliente.bean.master.sigep.bsb.correios.com.br/">
			<soapenv:Header/>
			<soapenv:Body>
			<cli:cancelarObjeto>				
				<idPlp>` + plp + `</idPlp>				
				<numeroEtiqueta>` + etiqueta + `</numeroEtiqueta>				
				<usuario>` + user + `</usuario>				
				<senha>` + senha + `</senha>
			</cli:cancelarObjeto>
			</soapenv:Body>
		</soapenv:Envelope>
	`)
	req, err := http.NewRequest("POST", Wsdl, strings.NewReader(payload))
	if err != nil {
		return err
	}
	http.DefaultClient.Transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	b, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return err
	}
	b, err = IsoUtf8(b)
	if strings.Contains(string(b), "faultstring") {
		respError := fault{}
		_ = xml.Unmarshal([]byte(b), &respError)
		return errors.New(respError.Body.Fault.FaultString)
	}
	sucesso := cancelarObjetoResponse{}
	_ = xml.Unmarshal([]byte(b), &sucesso)
	if !sucesso.Body.CancelarObjetoResponse.Retorno {
		return errors.New("erro desconhecido ao cancelar etiqueta")
	}
	return nil
}

//SolicitaEtiquetas faz a chamada ao SIGEPWEB e obtém uma faixa de etiquetas
func SolicitaEtiquetas(codigo string, identificador string, qtdEtiquetas int, user string, senha string) (string, error) {
	payload := fmt.Sprintf(`
		<x:Envelope
		xmlns:x="http://schemas.xmlsoap.org/soap/envelope/"
		xmlns:cli="http://cliente.bean.master.sigep.bsb.correios.com.br/">
		<x:Header/>
		<x:Body>
			<cli:solicitaEtiquetas>
				<tipoDestinatario>C</tipoDestinatario>
				<identificador>` + identificador + `</identificador>
				<idServico>` + codigo + `</idServico>
				<qtdEtiquetas>` + strconv.Itoa(qtdEtiquetas) + `</qtdEtiquetas>
				<usuario>` + user + `</usuario>
				<senha>` + senha + `</senha>
			</cli:solicitaEtiquetas>
		</x:Body>
	</x:Envelope>
	`)
	req, err := http.NewRequest("POST", Wsdl, strings.NewReader(payload))
	if err != nil {
		return "", err
	}
	http.DefaultClient.Transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	b, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return "", err
	}
	b, err = IsoUtf8(b)
	if strings.Contains(string(b), "faultstring") {
		respError := fault{}
		_ = xml.Unmarshal([]byte(b), &respError)
		return "", errors.New(respError.Body.Fault.FaultString)
	}
	faixa := solicitaEtiquetasResponse{}
	_ = xml.Unmarshal([]byte(b), &faixa)
	return faixa.Body.SolicitaEtiquetasResponse.FaixaEtiquetas, nil
}

//estrutura para conter o retorno do método geraDigitoVeiricadorEtiquetas
type geraDigitoVerificadorEtiquetasResponse struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    struct {
		XMLName                                xml.Name
		GeraDigitoVerificadorEtiquetasResponse struct {
			DigitoVerificador int `xml:"return"`
		} `xml:"geraDigitoVerificadorEtiquetasResponse"`
	} `xml:"Body"`
}

//GeraDigitoVerificadorEtiquetas faz a chamada ao SIGPEWEB e gera o dígito verificador de uma etiqueta
func GeraDigitoVerificadorEtiquetas(etiqueta string) (int, error) {
	payload := fmt.Sprintf(
		`<x:Envelope xmlns:x="http://schemas.xmlsoap.org/soap/envelope/" xmlns:cli="http://cliente.bean.master.sigep.bsb.correios.com.br/">
		 <x:Header/>
		 <x:Body>
			<cli:geraDigitoVerificadorEtiquetas>
				<etiquetas>` + etiqueta + `</etiquetas>
				<usuario> ` + User + `</usuario>
				<senha>` + Pass + `</senha>
			</cli:geraDigitoVerificadorEtiquetas>
		</x:Body>
		</x:Envelope>
		`)
	req, err := http.NewRequest("POST", Wsdl, strings.NewReader(payload))
	if err != nil {
		return 99, err
	}
	http.DefaultClient.Transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return 99, err
	}
	b, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return 99, err
	}
	b, err = IsoUtf8(b)
	if strings.Contains(string(b), "faultstring") {
		respError := fault{}
		_ = xml.Unmarshal([]byte(b), &respError)
		return 99, errors.New(respError.Body.Fault.FaultString)
	}
	digito := geraDigitoVerificadorEtiquetasResponse{}
	_ = xml.Unmarshal([]byte(b), &digito)
	return digito.Body.GeraDigitoVerificadorEtiquetasResponse.DigitoVerificador, nil
}

//estrutura para obter dados do contrato de um cliente
type buscaServicosResponse struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    struct {
		XMLName               xml.Name
		BuscaServicosResponse struct {
			Return []struct {
				Codigo    string `xml:"codigo"`
				ID        int    `xml:"id"`
				Descricao string `xml:"descricao"`
			} `xml:"return"`
		} `xml:"buscaServicosResponse"`
	}
}

//BuscaServicos faz a chamada ao SIGEPWEB e obtém dados de indentificao de um cliente
func BuscaServicos(contrato string, cartao string, usuario string, senha string) (buscaServicosResponse, error) {
	servicos := buscaServicosResponse{}
	payload := fmt.Sprintf(`
		<x:Envelope xmlns:x="http://schemas.xmlsoap.org/soap/envelope/" xmlns:cli="http://cliente.bean.master.sigep.bsb.correios.com.br/">
			<x:Header/>
				<x:Body>
					<cli:buscaServicos>
						<idContrato> ` + contrato + `</idContrato>
						<idCartaoPostagem>` + cartao + `</idCartaoPostagem>
						<usuario>` + usuario + `</usuario>
						<senha>` + senha + `</senha>
					</cli:buscaServicos>
			</x:Body>
		</x:Envelope>
	`)
	req, err := http.NewRequest("POST", Wsdl, strings.NewReader(payload))
	if err != nil {
		return servicos, err
	}
	http.DefaultClient.Transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return servicos, err
	}
	b, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return servicos, err
	}
	b, err = IsoUtf8(b)
	if strings.Contains(string(b), "faultstring") {
		respError := fault{}
		_ = xml.Unmarshal([]byte(b), &respError)
		return servicos, errors.New(respError.Body.Fault.FaultString)
	}

	_ = xml.Unmarshal([]byte(b), &servicos)
	return servicos, nil
}

//estrutura para conter os dados de um endereço a partir do CEP
type ConsultaCEPResponse struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    struct {
		ConsultaCEPResponse struct {
			Return struct {
				Bairro      string `xml:"bairro"`
				Cep         string `xml:"cep"`
				Cidade      string `xml:"cidade"`
				Complemento string `xml:"complemento2"`
				Endereco    string `xml:"end"`
				UF          string `xml:"uf"`
			} `xml:"return"`
		} `xml:"consultaCEPResponse"`
	} `xml:"Body"`
}

//ConsultaCEP faz a chamada ao SIGEPWEB e obtem o endereco correspondente a um CEP
func ConsultaCEP(cep string) (ConsultaCEPResponse, error) {
	endereco := ConsultaCEPResponse{}
	payload := fmt.Sprintf(
		`<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:cli="http://cliente.bean.master.sigep.bsb.correios.com.br/">
		<soapenv:Header/>
		<soapenv:Body>
			<cli:consultaCEP>
				<!--Optional:-->
				<cep>` + cep + `</cep>
			</cli:consultaCEP>
		</soapenv:Body>
		</soapenv:Envelope>`)
	req, err := http.NewRequest("POST", Wsdl, strings.NewReader(payload))
	if err != nil {
		return endereco, err
	}
	http.DefaultClient.Transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return endereco, err
	}
	b, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return endereco, err
	}
	b, err = IsoUtf8(b)
	if strings.Contains(string(b), "faultstring") {
		respError := fault{}
		_ = xml.Unmarshal([]byte(b), &respError)
		return endereco, errors.New(respError.Body.Fault.FaultString)
	}

	err = xml.Unmarshal([]byte(b), &endereco)
	if err != nil {
		return endereco, err
	}
	return endereco, nil
}

type solicitaPLPResponse struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    struct {
		XMLName             xml.Name
		SolicitaPLPResponse struct {
			XML string `xml:"return"`
		} `xml:"solicitaPLPResponse"`
	} `xml:"Body"`
}

//SolicitaPLP faz a chamada ao SIGEPWEB e obtem o xml de uma PLP
func SolicitaPLP(plp string, etiqueta string, usuario string, senha string) (string, error) {
	ret := solicitaPLPResponse{}
	payload := fmt.Sprintf(`
	<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:cli="http://cliente.bean.master.sigep.bsb.correios.com.br/">
		<soapenv:Header/>
			<soapenv:Body>
					<cli:solicitaPLP>
						<idPlpMaster>` + plp + `</idPlpMaster>
						<numEtiqueta>` + etiqueta + `</numEtiqueta>
						<usuario>` + usuario + `</usuario>
						<senha>` + senha + `</senha>
					</cli:solicitaPLP>
			</soapenv:Body>
		</soapenv:Envelope>`)
	req, err := http.NewRequest("POST", Wsdl, strings.NewReader(payload))
	if err != nil {
		return "", err
	}
	http.DefaultClient.Transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	b, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return "", err
	}
	b, err = IsoUtf8(b)
	if strings.Contains(string(b), "faultstring") {
		respError := fault{}
		_ = xml.Unmarshal([]byte(b), &respError)
		return "", errors.New(respError.Body.Fault.FaultString)
	}

	err = xml.Unmarshal([]byte(b), &ret)
	if err != nil {
		return "", err
	}
	return ret.Body.SolicitaPLPResponse.XML, nil
}

//estrutura para conter o numero de uma PLP
type fechaPlpVariosServicosResponse struct {
	XMLName xml.Name `xml:"Envelope"`
	Body    struct {
		XMLName                        xml.Name
		FechaPlpVariosServicosResponse struct {
			NumeroPLP string `xml:"return"`
		} `xml:"fechaPlpVariosServicosResponse"`
	}
}

//FechaPlpVariosServicos faz a chamada ao SIGPEWEB, fecha uma PLP
func FechaPlpVariosServicos(xmlPLP string, etiqueta string, etiquetaSemVerificador string, idPlpCliente string, cartao string, usuario string, senha string) (string, error) {
	xmlPLP = strings.Replace(xmlPLP, "XX000000000XX", etiqueta, 1)

	payload := fmt.Sprintf(
		`<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:cli="http://cliente.bean.master.sigep.bsb.correios.com.br/">
			<soapenv:Header/>
			<soapenv:Body>
				<cli:fechaPlpVariosServicos>
					<xml><![CDATA[` + xmlPLP + `]]></xml>
					<idPlpCliente>` + idPlpCliente + `</idPlpCliente>
					<cartaoPostagem>` + cartao + `</cartaoPostagem>
					<listaEtiquetas>` + etiquetaSemVerificador + `</listaEtiquetas>
					<usuario>` + usuario + `</usuario>
					<senha>` + senha + `</senha>
				</cli:fechaPlpVariosServicos>
			</soapenv:Body>
		</soapenv:Envelope>`)
	req, err := http.NewRequest("POST", Wsdl, strings.NewReader(payload))
	if err != nil {
		return "", err
	}
	http.DefaultClient.Transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	b, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return "", err
	}
	b, err = IsoUtf8(b)
	if strings.Contains(string(b), "faultstring") {
		respError := fault{}
		_ = xml.Unmarshal([]byte(b), &respError)
		return "", errors.New(respError.Body.Fault.FaultString)
	}
	plp := fechaPlpVariosServicosResponse{}

	err = xml.Unmarshal([]byte(b), &plp)
	if err != nil {
		return "", err
	}
	return plp.Body.FechaPlpVariosServicosResponse.NumeroPLP, nil
}
