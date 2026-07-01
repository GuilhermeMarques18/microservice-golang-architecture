package dto

import "fmt"

type ProcessRequestDTO struct {
	RequestID  string `json:"request_id"`
	Payload    []byte `json:"payload"`
	Complexity int32  `json:"complexity"`
}

func (dto *ProcessRequestDTO) Validate() error {
	if dto.RequestID == "" {
		return fmt.Errorf("Request_id é obrigatório")
	}

	if len(dto.RequestID) > 128 {
		return fmt.Errorf("Request_id não pode ultrapassar 128 caracteres")
	}

	if dto.Complexity <= 0 || dto.Complexity > 100_000 {
		return fmt.Errorf("Complexity deve estar entre 1 e 100000")
	}
	if len(dto.Payload) == 0 {
		return fmt.Errorf("Payload não pode ser vazio")
	}
	return nil
}
