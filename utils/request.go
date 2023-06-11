package request

import (
  "net/http"
  "struct"
)


func sendMailCode(data structs.RequestData) int {
  // Создаем клиент HTTP
	client := &http.Client{}

	// Создаем GET-запрос
	req, err := http.NewRequest("GET", data.Url, nil)
	if err != nil {
		return 404
	}

  req.Header.Add("Authorization", auth)
  req.Header.Add("Devid", devId)

	// Отправляем запрос
	resp, err := client.Do(req)
	if err != nil {
		return 404
	}
	defer resp.Body.Close()
	return resp.StatusCode
}
