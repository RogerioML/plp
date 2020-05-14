package plp

import (
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"

	"golang.org/x/text/encoding/charmap"
)

const (
	nuCliente = 2019020716
	//Versao indica a versão atual do módulo de manipulação de PLPs
	Versao      = "1.1.11"
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
)

func init() {
	erCartao = regexp.MustCompile(`^[0-9]{10}$`)
	erDr = regexp.MustCompile(`^[0-9]{2}$`)
	erCodAdm = regexp.MustCompile(`^[0-9]{8}$`)
	erCep = regexp.MustCompile(`^[0-9]{8}$`)
	erUf = regexp.MustCompile(`^[a-zA-Z]{2}$`)
	erTelefone = regexp.MustCompile(`^[0-9]*$`)
	erEmail = regexp.MustCompile(`^[a-zA-Z0-9.!#$%&’*+/=?^_{|}~-]+@[a-zA-Z0-9-]+(?:\.[a-zA-Z0-9-]+)*$`)
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

// Valida valida dados do remetente
func (p *Plp) Valida(idCorreios string) error {
	if !erCartao.MatchString(p.Plp.CartaoPostagem) {
		return ErrCartaoInvalido
	}
	if !erCartao.MatchString(p.Remetente.NumeroContrato) {
		return ErrContratoInvalido
	}
	if !erDr.MatchString(p.Remetente.NumeroDiretoria) {
		return ErrDrInvalido
	}
	if !erCodAdm.MatchString(p.Remetente.CodigoAdministrativo) {
		return ErrCodAdmInvalido
	}

	if p.Remetente.NomeRemetente.CData == "" || len(p.Remetente.NomeRemetente.CData) > 50 {
		return ErrNomeRemetente
	}
	if p.Remetente.LogradouroRemetente.CData == "" || len(p.Remetente.LogradouroRemetente.CData) > 50 {
		return ErrLogradouroRemetente
	}
	if p.Remetente.NumeroRemetente.CData == "" || len(p.Remetente.NumeroRemetente.CData) > 5 {
		return ErrNumeroRemetente
	}
	if len(p.Remetente.ComplementoRemetente.CData) > 30 {
		return ErrComplementoRemetente
	}
	if p.Remetente.BairroRemetente.CData == "" || len(p.Remetente.BairroRemetente.CData) > 30 {
		return ErrBairroRemetente
	}
	if !erCep.MatchString(p.Remetente.CepRemetente.CData) {
		return ErrCepRemetente
	}
	if p.Remetente.CidadeRemetente.CData == "" || len(p.Remetente.CidadeRemetente.CData) > 30 {
		return ErrCidadeRemetente
	}
	if !erUf.MatchString(p.Remetente.UfRemetente) {
		return ErrUfRemetente
	}
	if p.Remetente.TelefoneRemetente.CData != "" && !erTelefone.MatchString(p.Remetente.TelefoneRemetente.CData) {
		return ErrTelefoneRemetente
	}
	if p.Remetente.TelefoneRemetente.CData != "" && len(p.Remetente.TelefoneRemetente.CData) > 12 {
		return ErrTelefoneRemetente
	}
	if p.Remetente.EmailRemetente.CData != "" && !erEmail.MatchString(p.Remetente.EmailRemetente.CData) {
		return ErrEmailRemetente
	}
	if p.Remetente.CelularRemetente.CData != "" && !erTelefone.MatchString(p.Remetente.CelularRemetente.CData) {
		return ErrCelularRemetente
	}
	/*cartao, err := cc.Cartao(p.Plp.CartaoPostagem)
	if err != nil {
		return fmt.Errorf("plp valida 1: %s", err)
	}
	if cartao.Status != "01" {
		return fmt.Errorf("info: plp valida 2: cartão de postagem %q não está ativo", p.Plp.CartaoPostagem)
	}
	contratosPorNumeroDr, err := cc.Contratos("", p.Remetente.NumeroContrato, p.Remetente.NumeroDiretoria)
	if err != nil {
		return fmt.Errorf("plp valida 3: %s", err)
	}
	valido := false
	for _, c := range contratosPorNumeroDr {
		if c.Status == "A" {
			valido = true
			break
		}
	}
	if !valido {
		return fmt.Errorf(
			"info: plp valida 4: não existe contrato ativo com número %q e diretoria %q",
			p.Remetente.NumeroContrato,
			p.Remetente.NumeroDiretoria,
		)
	}
	*/
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

//FindXML busca uma PLP no banco tendo o numero da PLP como parâmetro de entrada
func FindXML(numero int) (string, error) {
	query := `
		SELECT	PLP_IN_STATUS, PLP_TX_XML, PLP_TX_XML_RETORNO
		FROM NEP_PRE_LISTA_POSTAGEM
		WHERE PLP_NU = :numero	
	`
	var (
		plpInStatus  int
		txXML        string
		txXMLRetorno string
	)
	row := db.QueryRow(query, numero)
	if err := row.Scan(
		&plpInStatus,
		&txXML,
		&txXMLRetorno,
	); err != nil {
		return "", fmt.Errorf("plp PlpFindXML 1: %s", err)
		//PLP não encontrada
		//return "", nil
	}

	if txXMLRetorno != "" {
		return txXMLRetorno, nil
	}
	return txXML, nil
}

//Disponibilidade retorna a quantidade de dias que uma PLP fica disponível para postagem
func Disponibilidade() (int, error) {
	query := `
		SELECT PRM_TE_VALOR FROM NEP_PARAMETRO_PLP WHERE PRM_TX_DESCRICAO = 'DisponibilidadePLP'
	`
	var dias int
	row := db.QueryRow(query)
	if err := row.Scan(&dias); err != nil {
		return 0, fmt.Errorf("plp disponibilidade: %s", err)
	}
	return dias, nil
}

//AtualizaXML grava alterações realizadas na PLP diretamente no Banco de Dados
func (p *Plp) AtualizaXML() error {
	tx, err := db.Begin()
	docXML, err := p.XML()
	if err != nil {
		return fmt.Errorf("plp atualiza xml 1: %s", err)
	}
	query := `UPDATE NEP_PRE_LISTA_POSTAGEM SET PLP_TX_XML = :docXML WHERE PLP_NU = :id`
	_, err = tx.Exec(query, docXML, p.Plp.IDPlp)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return rollbackErr
		}
		return fmt.Errorf("plp atualizaXML 2: oracle: %s", err)
	}
	tx.Commit()
	return nil
}

//ObtemListaPLPSIntactas retorna uma lista de PLPS de um determinado contrato que não sofreram manipulações em nenhum sistema
func ObtemListaPLPSIntactas(contrato string, dias int) ([]int, error) {
	var plps []int
	query := `
		SELECT   /*+ PARALLEL(32) */
		NEP_PreLstPtg.PLP_nu		
		FROM          NEP_Pre_Lista_Postagem NEP_PreLstPtg
		JOIN Vw_F5536B              VwF5536B
		ON (  VwF5536B.ITZIACPJDE
		= NEP_PreLstPtg.CPZIACPJDE)
		WHERE    (    (  trim(VwF5536B.ITDMCT)
		= :contrato                                 ) -- <---<<< Informe o número do contrato.
		AND (        NEP_PreLstPtg.PLP_in_Incorporada
		<>      'S'                                    )
		AND NOT EXISTS(SELECT 1
		FROM   NEP_Objeto_Postal NEP_ObjPst
			WHERE  (    (  NEP_ObjPst.PLP_nu
							= NEP_PreLstPtg.PLP_nu)
					AND (  NEP_ObjPst.Obj_in_Status
							= 1                   ))) -- 1 - Postado.
		AND (        NEP_PreLstPtg.PLP_dt_Postagem
		BETWEEN (  TRUNC(SYSDATE)
			- :dias)
		AND     (TRUNC(SYSDATE) + 86399/86400) ))
		GROUP BY  NEP_PreLstPtg.PLP_nu`

	rows, err := db.Query(query, contrato, dias)
	if err != nil {
		return plps, err
	}
	defer rows.Close()
	for rows.Next() {
		var plp int
		err := rows.Scan(&plp)
		if err != nil {
			return plps, fmt.Errorf("erro ao encontrar plps do contrato: %s", err)
		}
		plps = append(plps, plp)
	}
	if len(plps) == 0 {
		return plps, errors.New("nenhuma plp intacta foi localizada para o contrato: " + contrato)
	}
	return plps, nil

}

//IsoUtf8 converte de ISO para UTF-8
func IsoUtf8(b []byte) ([]byte, error) {
	r := charmap.ISO8859_1.NewDecoder().Reader(strings.NewReader(string(b)))
	return ioutil.ReadAll(r)
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

//FechaPlpVariosServicos faz a chamada ao SIGPEWEB, fecha uma PLP
func FechaPlpVariosServicos(wsdl string) (string, error) {
	payload := fmt.Sprintf(
		`<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:cli="http://cliente.bean.master.sigep.bsb.correios.com.br/">
			<soapenv:Header/>
			<soapenv:Body>
				<cli:fechaPlpVariosServicos>
					<!--Optional:-->
					<xml><![CDATA[<?ml version="1.0" encoding="ISO-8859-1"?><correioslog><tipo_arquivo>Postagem</tipo_arquivo><versao_arquivo>2.3</versao_arquivo><plp><id_plp /><valor_global/><mcu_unidade_postagem/><nome_unidade_postagem/><cartao_postagem>0068600275</cartao_postagem></plp><remetente><numero_contrato>9912208555</numero_contrato><numero_diretoria>10</numero_diretoria><codigo_administrativo>08082650</codigo_administrativo><nome_remetente>Monitor de Fechamento de PLP</nome_remetente><logradouro_remetente>SNQ Quadra 1 Bloco A 2º SS</logradouro_remetente><numero_remetente>0</numero_remetente><complemento_remetente/><bairro_remetente>Asa Norte</bairro_remetente><cep_remetente>70002900</cep_remetente><cidade_remetente>Brasília</cidade_remetente><uf_remetente>DF</uf_remetente><telefone_remetente>6121416129</telefone_remetente><fax_remetente/><email_remetente/></remetente><forma_pagamento/><objeto_postal><numero_etiqueta>OH566322999BR</numero_etiqueta><codigo_objeto_cliente/><codigo_servico_postagem>04162</codigo_servico_postagem><cubagem>0,0000</cubagem><peso>800</peso><rt1/><rt2/><destinatario><nome_destinatario>Correios DETEC</nome_destinatario><telefone_destinatario>6121416129</telefone_destinatario><celular_destinatario/><email_destinatario/><logradouro_destinatario>SNN Quadra 1 Bloco A</logradouro_destinatario><complemento_destinatario/><numero_end_destinatario>0</numero_end_destinatario></destinatario><nacional><bairro_destinatario>Asa Norte</bairro_destinatario><cidade_destinatario>Brasília</cidade_destinatario><uf_destinatario>DF</uf_destinatario><cep_destinatario>70002900</cep_destinatario><codigo_usuario_postal/><centro_custo_cliente/><numero_nota_fiscal>1234567</numero_nota_fiscal><serie_nota_fiscal/><valor_nota_fiscal/><natureza_nota_fiscal/><descricao_objeto/><valor_a_cobrar>0,0</valor_a_cobrar></nacional><servico_adicional><codigo_servico_adicional>025</codigo_servico_adicional><valor_declarado/></servico_adicional><dimensao_objeto><tipo_objeto>002</tipo_objeto><dimensao_altura>50</dimensao_altura><dimensao_largura>30</dimensao_largura><dimensao_comprimento>40</dimensao_comprimento><dimensao_diametro>0</dimensao_diametro></dimensao_objeto><data_postagem_sara/><status_processamento>0</status_processamento><numero_comprovante_postagem/><valor_cobrado/></objeto_postal></correioslog>]]></xml>
					<!--Optional:-->
					<idPlpCliente>54321</idPlpCliente>
					<!--Optional:-->
					<cartaoPostagem>0068600275</cartaoPostagem>
					<!--Zero or more repetitions:-->
					<listaEtiquetas>OH56632299BR</listaEtiquetas>
					<!--Optional:-->
					<usuario>gati</usuario>
					<!--Optional:-->
					<senha>lbqhj</senha>
				</cli:fechaPlpVariosServicos>
			</soapenv:Body>
		</soapenv:Envelope>`)
	req, err := http.NewRequest("POST", wsdl, strings.NewReader(payload))
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
	_ = xml.Unmarshal([]byte(b), &plp)
	return plp.Body.FechaPlpVariosServicosResponse.NumeroPLP, nil
}

func removePlp(plpNu string, db *sql.DB) error {
	//remove os objetos da PLP do banco
	now := time.Now()
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("erro conectar no banco para excluir plp: %s", err)
	}
	query := "DELETE FROM NEP_OBJETO_POSTAL WHERE PLP_NU = :plp"
	_, err = tx.Exec(query, plpNu)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("erro ao excluir plp: %s", err)
		}
		return fmt.Errorf("erro ao excluir plp: %s", err)
	}
	//remove a PLP do banco
	query = "DELETE FROM NEP_PRE_LISTA_POSTAGEM WHERE PLP_NU = :plp"
	_, err = tx.Exec(query, plpNu)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("erro ao excluir plp: %s", err)
		}
		return fmt.Errorf("erro ao excluir plp: %s", err)
	}
	fmt.Printf("%s plp %s excluida em %.3f segundos\n", now.Format(LayoutMysql), plpNu, time.Since(now).Seconds())
	return tx.Commit()
}

func removePlpPorEtiqueta(etiqueta string, db *sql.DB) error {
	//remove os objetos da PLP do banco
	now := time.Now()
	tx, err := db.Begin()
	var plpNu int64

	row := db.QueryRow("SELECT PLP_NU FROM NEP_OBJETO_POSTAL WHERE obj_nu_etiqueta = :etiqueta", etiqueta)

	err = row.Scan(&plpNu)

	if err != nil {
		return fmt.Errorf("erro conectar no banco para obter plp a excluir: %s", err)
	}
	query := `DELETE FROM NEP_OBJETO_POSTAL WHERE PLP_NU = :plp`
	_, err = tx.Exec(query, plpNu)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("erro ao excluir plp: %s", err)
		}
		return fmt.Errorf("erro ao excluir plp: %s", err)
	}
	//remove a PLP do banco
	query = `DELETE FROM NEP_PRE_LISTA_POSTAGEM WHERE PLP_NU =	:plp`
	_, err = tx.Exec(query, plpNu)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("erro ao excluir plp: %s", err)
		}
		return fmt.Errorf("erro ao excluir plp: %s", err)
	}
	fmt.Printf("%s plp %s excluida em %.3f segundos\n", now.Format(LayoutMysql), plpNu, time.Since(now).Seconds())
	return tx.Commit()
}