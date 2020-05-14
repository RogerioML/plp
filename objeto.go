package plp

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

//Estados possíveis para os objetos postais
const (
	EstadoObjetoEmAberto              = 0
	EstadoObjetoPostado               = 1
	EstadoObjetoCancelado             = 2
	EstadoObjetoEmConferencia         = 3
	EstadoObjetoConferido             = 4
	EstadoObjetoConferenciaFinalizada = 5
	EstadoObjetoRemovidoDaConferencia = 6
)

//Relação de erros possíveis para os objetos postais
var (
	ErrObjetoPostado = errors.New("negocio: objeto ja foi postado")
)

// CodigoServicoAdicional complemento da estrutura Objeto XML
type CodigoServicoAdicional struct {
	CodigoServicoAdicional []string `xml:"codigo_servico_adicional"`
	ValorDeclarado         string   `xml:"valor_declarado"`
	EnderecoVizinho        struct {
		CData string `xml:",cdata"`
	} `xml:"endereco_vizinho"`
}

// Objeto estrutura que converte para o XML
type Objeto struct {
	NumeroEtiqueta string `xml:"numero_etiqueta"`
	//PlpNu                 int       //`xml:"plp"`
	Inclusao              time.Time `xml:"-"`
	StatusTabela          int       `xml:"-"`
	CodigoObjetoCliente   string    `xml:"codigo_objeto_cliente"`
	CodigoServicoPostagem string    `xml:"codigo_servico_postagem"`
	Cubagem               string    `xml:"cubagem"`
	Peso                  int       `xml:"peso"`
	Rt1                   string    `xml:"rt1"`
	Rt2                   string    `xml:"rt2"`
	RestricaoANAC         string    `xml:"restricao_anac"`
	Destinatario          struct {
		NomeDestinatario struct {
			CData string `xml:",cdata"`
		} `xml:"nome_destinatario"`
		TelefoneDestinatario struct {
			CData string `xml:",cdata"`
		} `xml:"telefone_destinatario"`
		CelularDestinatario struct {
			CData string `xml:",cdata"`
		} `xml:"celular_destinatario"`
		EmailDestinatario struct {
			CData string `xml:",cdata"`
		} `xml:"email_destinatario"`
		LogradouroDestinatario struct {
			CData string `xml:",cdata"`
		} `xml:"logradouro_destinatario"`
		ComplementoDestinatario struct {
			CData string `xml:",cdata"`
		} `xml:"complemento_destinatario"`
		NumeroEndDestinatario struct {
			CData string `xml:",cdata"`
		} `xml:"numero_end_destinatario"`
		CpfCnpjDestinatario string `xml:"cpf_cnpj_destinatario"`
	} `xml:"destinatario"`
	Nacional struct {
		BairroDestinatario struct {
			CData string `xml:",cdata"`
		} `xml:"bairro_destinatario"`
		CidadeDestinatario struct {
			CData string `xml:",cdata"`
		} `xml:"cidade_destinatario"`
		UfDestinatario  string `xml:"uf_destinatario"`
		CepDestinatario struct {
			CData string `xml:",cdata"`
		} `xml:"cep_destinatario"`
		CodigoUsuarioPostal string `xml:"codigo_usuario_postal"`
		CentroCustoCliente  string `xml:"centro_custo_cliente"`
		NumeroNotaFiscal    string `xml:"numero_nota_fiscal"`
		SerieNotaFiscal     string `xml:"serie_nota_fiscal"`
		ValorNotaFiscal     string `xml:"valor_nota_fiscal"`
		NaturezaNotaFiscal  string `xml:"natureza_nota_fiscal"`
		DescricaoObjeto     struct {
			CData string `xml:",cdata"`
		} `xml:"descricao_objeto"`
		ValorACobrar string `xml:"valor_a_cobrar"`
	} `xml:"nacional"`
	ServicoAdicional []CodigoServicoAdicional `xml:"servico_adicional"`
	Dimensoes        struct {
		Tipo        string `xml:"tipo_objeto"`
		Altura      string `xml:"dimensao_altura"`
		Largura     string `xml:"dimensao_largura"`
		Comprimento string `xml:"dimensao_comprimento"`
		Diametro    string `xml:"dimensao_diametro"`
	} `xml:"dimensao_objeto"`
	DataPostagemSara          string  `xml:"data_postagem_sara"`
	StatusProcessamento       string  `xml:"status_processamento"`
	NumeroComprovantePostagem string  `xml:"numero_comprovante_postagem"`
	ValorCobrado              float64 `xml:"valor_cobrado"`
}

// CodigoServicoAdicionalJSON complemento da estrutura Objeto JSON
type CodigoServicoAdicionalJSON struct {
	CodigoServicoAdicional string `json:"codigo_servico_adicional"`
}

// ObjetoJSON estrutura do objeto com JSON
type ObjetoJSON struct {
	Etiqueta      string           `json:"etiqueta"`
	Mcu           string           `json:"mcu"`
	DataCriacao   time.Time        `json:"data_criacao"`
	Status        int              `json:"status"`
	CodigoServico string           `json:"codigo_servico"`
	Cubagem       string           `json:"cubagem"`
	Peso          int              `json:"peso"`
	Plp           PlpRemetenteJSON `json:"plp"`

	Destinatario struct {
		Nome                string `json:"nome"`
		Telefone            string `json:"telefone"`
		Celular             string `json:"celular_destinatario"`
		Email               string `json:"email"`
		Logradouro          string `json:"logradouro"`
		Complemento         string `json:"complemento"`
		Numero              string `json:"numero"`
		Bairro              string `json:"bairro"`
		Cidade              string `json:"cidade"`
		UF                  string `json:"uf"`
		Cep                 string `json:"cep"`
		CodigoUsuarioPostal string `json:"-"`
		CentroCustoCliente  string `json:"-"`
		NotaFiscal          string `json:"numero_nota_fiscal"`
		SerieNotaFiscal     string `json:"serie_nota_fiscal"`
		ValorNotaFiscal     string `json:"valor_nota_fiscal"`
		NaturezaNotaFiscal  string `json:"natureza_nota_fiscal"`
		Descricao           string `json:"descricao"`
		ValorACobrar        string `json:"valor_a_cobrar"`
	} `json:"destinatario"`
	ServicoAdicional []CodigoServicoAdicionalJSON `json:"servico_adicional"`
	Dimensoes        struct {
		Tipo        string `json:"tipo"`
		Altura      string `json:"altura"`
		Largura     string `json:"largura"`
		Comprimento string `json:"comprimento"`
		Diametro    string `json:"diametro"`
	} `json:"dimensoes"`
	DataPostagem              string  `json:"data_postagem"`
	NumeroComprovantePostagem string  `json:"comprovante_postagem"`
	ValorCobrado              float64 `json:"valor_cobrado"`
}

var (
	erEtiqueta *regexp.Regexp
	// ErrPLPNaoRascunho unidade do MCU informado não encontrada
	ErrPLPNaoRascunho = errors.New("objeto plp não rascunho: a plp não permite mais alteração")
)

func init() {
	erEtiqueta = regexp.MustCompile(`[A-Z]{2}[0-9]{8}[ ]*[A-Z]{2}`)
}

//TrocaServico efetua a troca do serviço dentro do XML da PLP conforme critérios da nova política comercial
func (o *Objeto) TrocaServico(pacoteOrigem string, pacoteDestino string) error {
	switch pacoteOrigem {
	case "2.0":
		switch pacoteDestino {
		case "bronze":
			switch o.CodigoServicoPostagem {
			case "04537":
				o.CodigoServicoPostagem = "03042"
				break
			case "04553":
				o.CodigoServicoPostagem = "03050"
				break
			case "04596":
				o.CodigoServicoPostagem = "03085"
				break
			case "04618":
				o.CodigoServicoPostagem = "03107"
				break
			case "40215":
				o.CodigoServicoPostagem = "04790"
				break
			case "40169":
				o.CodigoServicoPostagem = "04782"
				break
			case "40290":
				o.CodigoServicoPostagem = "04804"
				break
			}

		case "prata", "ouro", "platinum", "diamante", "infinite":
			switch o.CodigoServicoPostagem {
			case "04537":
				o.CodigoServicoPostagem = "03212"
				break
			case "04553":
				o.CodigoServicoPostagem = "03220"
				break
			case "04596":
				o.CodigoServicoPostagem = "03298"
				break
			case "04618":
				o.CodigoServicoPostagem = "03328"
				break
			case "40215":
				o.CodigoServicoPostagem = "03158"
				break
			case "40169":
				o.CodigoServicoPostagem = "03140"
				break
			case "40290":
				o.CodigoServicoPostagem = "03204"
				break
			}
		}
	case "2.1", "2.2", "2.3", "2.4", "2.5", "2.6", "2.7", "2.8", "2.9":
		switch pacoteDestino {
		case "bronze":
			switch o.CodigoServicoPostagem {
			case "40169":
				o.CodigoServicoPostagem = "04782"
				break
			case "40215":
				o.CodigoServicoPostagem = "04790"
				break
			case "40290":
				o.CodigoServicoPostagem = "04804"
				break
			case "04138":
				o.CodigoServicoPostagem = "03042"
				break
			case "04162":
				o.CodigoServicoPostagem = "03050"
				break
			case "04669":
				o.CodigoServicoPostagem = "03085"
				break
			case "04693":
				o.CodigoServicoPostagem = "03107"
				break
			}
		case "prata", "ouro", "platinum", "diamante", "infinite":
			switch o.CodigoServicoPostagem {
			case "40169":
				o.CodigoServicoPostagem = "03140"
				break
			case "40215":
				o.CodigoServicoPostagem = "03158"
				break
			case "40290":
				o.CodigoServicoPostagem = "03204"
				break
			case "04138":
				o.CodigoServicoPostagem = "03212"
				break
			case "04162":
				o.CodigoServicoPostagem = "03220"
				break
			case "04669":
				o.CodigoServicoPostagem = "03298"
				break
			case "04693":
				o.CodigoServicoPostagem = "03328"
				break
			}
		}
	case "5.1", "5.2", "5.3", "5.4", "5.5", "5.6", "5.7":
		switch pacoteDestino {
		case "bronze":
			switch o.CodigoServicoPostagem {
			case "40169":
				o.CodigoServicoPostagem = "04782"
				break
			case "40215":
				o.CodigoServicoPostagem = "04790"
				break
			case "40290":
				o.CodigoServicoPostagem = "04804"
				break
			case "04138":
				o.CodigoServicoPostagem = "03042"
				break
			case "04162":
				o.CodigoServicoPostagem = "03050"
				break
			case "04669":
				o.CodigoServicoPostagem = "03085"
				break
			case "04693":
				o.CodigoServicoPostagem = "03107"
				break
			case "04316":
				o.CodigoServicoPostagem = "03050"
				break
			case "04812":
				o.CodigoServicoPostagem = "03085"
				break
			}
		case "prata", "ouro", "platinum":
			switch o.CodigoServicoPostagem {
			case "40169":
				o.CodigoServicoPostagem = "03140"
				break
			case "40215":
				o.CodigoServicoPostagem = "03158"
				break
			case "40290":
				o.CodigoServicoPostagem = "03204"
				break
			case "04138":
				o.CodigoServicoPostagem = "03212"
				break
			case "04162":
				o.CodigoServicoPostagem = "03220"
				break
			case "04669":
				o.CodigoServicoPostagem = "03298"
				break
			case "04693":
				o.CodigoServicoPostagem = "03328"
				break
			case "04316":
				o.CodigoServicoPostagem = "03220"
				break
			case "04812":
				o.CodigoServicoPostagem = "03298"
				break
			}
		case "diamante", "infinite":
			switch o.CodigoServicoPostagem {
			case "40169":
				o.CodigoServicoPostagem = "03140"
				break
			case "40215":
				o.CodigoServicoPostagem = "03158"
				break
			case "40290":
				o.CodigoServicoPostagem = "03204"
				break
			case "04138":
				o.CodigoServicoPostagem = "03212"
				break
			case "04162":
				o.CodigoServicoPostagem = "03220"
				break
			case "04669":
				o.CodigoServicoPostagem = "03298"
				break
			case "04693":
				o.CodigoServicoPostagem = "03328"
				break
			case "04316":
				o.CodigoServicoPostagem = "03280"
				break
			case "04812":
				o.CodigoServicoPostagem = "03336"
				break
			}

		}
	case "bronze":
		switch o.CodigoServicoPostagem {
		case "04782":
			o.CodigoServicoPostagem = "03140"
			break
		case "04790":
			o.CodigoServicoPostagem = "03158"
			break
		case "04804":
			o.CodigoServicoPostagem = "03204"
			break
		case "03042":
			o.CodigoServicoPostagem = "03212"
			break
		case "03052":
			o.CodigoServicoPostagem = "03220"
			break
		case "03085":
			o.CodigoServicoPostagem = "03298"
			break
		case "03107":
			o.CodigoServicoPostagem = "03328"
			break
		}
	case "diamante", "infinite":
		switch o.CodigoServicoPostagem {
		case "03140":
			o.CodigoServicoPostagem = "04782"
			break
		case "03158":
			o.CodigoServicoPostagem = "04790"
			break
		case "03204":
			o.CodigoServicoPostagem = "04804"
			break
		case "03212":
			o.CodigoServicoPostagem = "03042"
			break
		case "03220":
			o.CodigoServicoPostagem = "03050"
			break
		case "03298":
			o.CodigoServicoPostagem = "03085"
			break
		case "03328":
			o.CodigoServicoPostagem = "03107"
			break
		case "03282":
			o.CodigoServicoPostagem = "03050"
			break
		case "03336":
			o.CodigoServicoPostagem = "03085"
			break
		}
	}
	return nil
}

func (o *ObjetoJSON) isPostado() bool {
	var status int
	qry := `select OBJ_IN_STATUS from NEP_OBJETO_POSTAL where obj_nu_etiqueta = :plpnu`
	if err := db.QueryRow(qry, o.Etiqueta).Scan(&status); err != nil {
		return false
	}
	return status == EstadoObjetoPostado
}

//Valida os objetos postais
func (o *ObjetoJSON) Valida() error {
	if o.isPostado() {
		return ErrObjetoPostado
	}
	return nil
}

// Insere objeto na estrutura xml
/*
func (o *Objeto) Insere(tx *sql.Tx) error {
	var (
		temTx bool
		err   error
	)
	if tx != nil {
		temTx = true
	} else {
		tx, err = db.Begin()
		if err != nil {
			return err
		}
	}
	query := `
		INSERT INTO NEP_OBJETO_POSTAL
		(OBJ_NU_ETIQUETA, PLP_NU, OBJ_DT_INCLUSAO)
		VALUES
		(:etiqueta, :numero, to_date(:inclusao,'ddmmyyyyhh24miss'))
	`
	etiqueta := o.NumeroEtiqueta
	if len(etiqueta) == 13 {
		etiqueta = etiqueta[:10] + etiqueta[11:]
	}
	_, err = tx.Exec(query, etiqueta, o.PlpNu, toDate(o.Inclusao))
	if err != nil {
		return err
	}
	if !temTx {
		return tx.Commit()
	}
	return nil
}
*/
// EtiquetaDV func para criar o dígito verificados
func EtiquetaDV(numero string) (string, error) {
	numEti := strings.Replace(numero, " ", "", -1)
	if len(numEti) == 13 {
		return numEti, nil
	}
	if !erEtiqueta.MatchString(numEti) {
		return "", errors.New(fmt.Sprintf("etiqueta %s inválida", numEti))
	}
	numeros := numEti[2:10]
	multiplicadores := [...]int{8, 6, 4, 2, 3, 5, 9, 7}
	soma := 0
	for i := 0; i < 8; i++ {
		numero, err := strconv.Atoi(numeros[i : i+1])
		if err != nil {
			return "", errors.New(fmt.Sprintf("etiqueta %s inválida", numEti))
		}
		soma += numero * multiplicadores[i]
	}
	resto := soma % 11
	var dv string
	switch resto {
	case 0:
		dv = "5"
	case 1:
		dv = "0"
	default:
		dv = strconv.Itoa(11 - resto)
	}
	return fmt.Sprintf("%v%v%v", numEti[:10], dv, numEti[10:]), nil
}

//IntervaloEtiquetas recebe uma string com um intervalo de etiquetas no formato "SZ46641024 BR,SZ46642023 BR"
//proveniente do método solicitaEtiquetas do SIGEP Master e devolve um slice com as etiquetas correspondentes
//com o dígito verificador
func IntervaloEtiquetas(intervalo string) ([]string, error) {
	partes := strings.Split(intervalo, ",")
	ini, fim := partes[0], partes[1]
	if !erEtiqueta.MatchString(ini) {
		return nil, errors.New(fmt.Sprintf("intervaloetiquetas: etiqueta %s inválida", ini))
	}
	if !erEtiqueta.MatchString(fim) {
		return nil, errors.New(fmt.Sprintf("intervaloetiquetas: etiqueta %s inválida", fim))
	}
	ini = strings.Replace(ini, " ", "", 1)
	fim = strings.Replace(fim, " ", "", 1)
	preIni := ini[0:2]
	sufIni := ini[len(ini)-2:]
	numIni := ini[2 : len(ini)-2]
	preFim := fim[0:2]
	sufFim := fim[len(fim)-2:]
	numFim := fim[2 : len(fim)-2]
	if preIni != preFim {
		return nil, errors.New(fmt.Sprintf("intervaloetiquetas: prefixos inicial e final não coincidem"))
	}
	if sufIni != sufFim {
		return nil, errors.New(fmt.Sprintf("intervaloetiquetas: sufixos inicial e final não coincidem"))
	}
	i, _ := strconv.Atoi(numIni)
	f, _ := strconv.Atoi(numFim)
	if f < i {
		return nil, errors.New(fmt.Sprintf("intervaloetiquetas: inicial deve ser menor que final"))
	}
	etiquetas := make([]string, 0, f-i+1)
	for i <= f {
		e, err := EtiquetaDV(fmt.Sprintf("%v%v%v", preIni, i, sufIni))
		if err != nil {
			return nil, errors.New(fmt.Sprintf("intervaloetiquetas: %s", err))
		}
		etiquetas = append(etiquetas, e)
		i++
	}
	return etiquetas, nil
}
