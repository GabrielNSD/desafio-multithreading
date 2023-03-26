package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type ViaCEP struct {
	Cep         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	Uf          string `json:"uf"`
	Ibge        string `json:"ibge"`
	Gia         string `json:"gia"`
	Ddd         string `json:"ddd"`
	Siafi       string `json:"siafi"`
}

type APICEP struct {
	Code       string `json:"code"`
	State      string `json:"state"`
	City       string `json:"city"`
	District   string `json:"district"`
	Address    string `json:"address"`
	Status     int    `json:"status"`
	Ok         bool   `json:"ok"`
	StatusText string `json:"statusText"`
}

type Result struct {
	API        string
	CEP        string
	Localidade string
	UF         string
	Logradouro string
}

func main() {
	c1 := make(chan Result)
	c2 := make(chan Result)

	for _, cep := range os.Args[1:] {
		cepLength := len(cep)
		cepFormatted := cep
		if cepLength == 8 {
			cepFormatted = cep[:5] + "-" + cep[5:]
		}

		go func() {
			viaCep, err := BuscaCEPVIACEP(cepFormatted)
			if err != nil {
				close(c1)
			}
			result := Result{
				API:        "ViaCEP",
				CEP:        viaCep.Cep,
				Localidade: viaCep.Localidade,
				UF:         viaCep.Uf,
				Logradouro: viaCep.Logradouro,
			}
			c1 <- result
		}()

		go func() {
			apiCep, err := BuscaCEPAPICEP(cepFormatted)
			if err != nil {
				close(c2)
			}
			result := Result{
				API:        "APICEP",
				CEP:        apiCep.Code,
				Localidade: apiCep.City,
				UF:         apiCep.State,
				Logradouro: apiCep.Address,
			}
			c2 <- result
		}()

		select {
		case r1 := <-c1:
			fmt.Printf("Received from ViaCEP\nCEP: %s\n Localidade: %s\n UF: %s\n Logradouro: %s", r1.CEP, r1.Localidade, r1.UF, r1.Logradouro)

		case r2 := <-c2:
			fmt.Printf("Received from APICEP\nCEP: %s\n Localidade: %s\n UF: %s\n Logradouro: %s", r2.CEP, r2.Localidade, r2.UF, r2.Logradouro)

		case <-time.After(1 * time.Second):
			fmt.Println("Timeout")
		}

	}

}

func BuscaCEPVIACEP(cep string) (*ViaCEP, error) {
	resp, err := http.Get("https://viacep.com.br/ws/" + cep + "/json/")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var c ViaCEP
	err = json.Unmarshal(body, &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func BuscaCEPAPICEP(cep string) (*APICEP, error) {
	resp, err := http.Get("https://cdn.apicep.com/file/apicep/" + cep + ".json")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var c APICEP
	err = json.Unmarshal(body, &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}
