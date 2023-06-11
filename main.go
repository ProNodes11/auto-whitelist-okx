package main

import (
  "io"
  "os"
	"fmt"
  "log"
  "time"
  "bytes"
  "bufio"
  "regexp"
  // "syscall"
  "strings"
  "net/http"
  "io/ioutil"
  // "os/signal"
  "github.com/joho/godotenv"
	"github.com/pquerna/otp/totp"
  "encoding/json"
  "github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
  "github.com/emersion/go-message/mail"
  "struct"
)

type AddressInfo struct {
	Address      string `json:"address"`
	ValidateName string `json:"validateName"`
}


var (
  maxAttempts = 5
	attempt     = 0
  status int
  envVars = make(map[string]string)
  wallets = []string{}
  senderEmail = "noreply@mailer"
  timeFormat  = "2006-01-02 15:04:05"
)

func main() {
  envVars = setEnvVars()  // Выгружаем все переменные которые задали
  wallets = setWallets() // Получаем все кошельки
  myStryct := structs.RequestData{
    Url: "pipa",
  }
  fmt.Println(myStryct)
  // start()  // Запускаем добавление кошелей
}

func setEnvVars() map[string]string {
  err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Ошибка при загрузке файла конфигурации", err)
	}

  envVars := make(map[string]string)

	// Добавление ключа и значения в map
	envVars["OKX_AUTORIZATION"]   = os.Getenv("OKX_AUTORIZATION")
  envVars["OKX_DEVID"]          = os.Getenv("OKX_DEVID")
  envVars["AUTENTIFICATOR_KEY"] = os.Getenv("AUTENTIFICATOR_KEY")
  envVars["IMAP_SERVER"]        = os.Getenv("IMAP_SERVER")
  envVars["EMAIL_ADDRESS"]      = os.Getenv("EMAIL_ADDRESS")
  envVars["EMAIL_PASSWORD"]     = os.Getenv("EMAIL_PASSWORD")

	// Вывод значений map
	// for key, value := range envVars {
	// 	log.Printf("%s = %s\n", key, value)
	// }
  return envVars
}

func setWallets() []string {
  file, err := os.Open("wallet.txt")
	if err != nil {
		log.Fatal("Ошибка открытия файла с кошельками: ", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
  var lines []string

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		log.Fatal("Ошибка чтения файла с кошельками: ", err)
	}
  return lines
}

func start() {
  chunkSize := 20
	totalWallets := len(wallets)
	numChunks := (totalWallets + chunkSize - 1) / chunkSize

	for i := 0; i < numChunks; i++ {
		start := i * chunkSize
		end := (i + 1) * chunkSize
		if end > totalWallets {
			end = totalWallets
		}
		chunk := wallets[start:end]
		processChunk(chunk)
	}
}

func processChunk(chunk []string) {
  log.Printf("Начинаем добавлять часть кошельков")
  addressStr := addressStrGen(chunk)

  // Отправляем запрос с инициализацией на почту
  for attempt <= maxAttempts {
    status = initAddAddress("start", "", "", envVars["OKX_AUTORIZATION"], envVars["OKX_DEVID"], addressStr)
    if status == 200 {
        log.Printf("Отправка запроса на биржу удалась")
        break
    }
    attempt++
    if attempt > maxAttempts {
        log.Printf("Достигнуто максимальное количество попыток. Адреса не были добавлены")
        attempt = 0
        break
    }
    log.Printf("Отправка подтверждения на биржу не удалось. Повторная попытка %d...\n", attempt)
    time.Sleep(time.Second * 10)
  }
  // Отправляем запрос на почту
  for attempt <= maxAttempts {
    status = sendMailCode(envVars["OKX_AUTORIZATION"], envVars["OKX_DEVID"], chunk)
    if status == 200 {
        log.Printf("Отправка подтверджения на почту удалась")
        break
    }
    attempt++
    if attempt > maxAttempts {
        log.Printf("Достигнуто максимальное количество попыток. Адреса не были добавлены")
        attempt = 0
        break
    }
    log.Printf("Отправка подтверждения на почту не удалось. Повторная попытка %d...\n", attempt)
    time.Sleep(time.Second * 10)
  }
  // Получаем код с почты
  emailCode := imapClient(envVars["IMAP_SERVER"], envVars["EMAIL_ADDRESS"], envVars["EMAIL_PASSWORD"])
  // Получаем код аутентификатора и отправляем запрос на биржу
  for attempt <= maxAttempts {
    AutentificatorCode := authCode(envVars["AUTENTIFICATOR_KEY"])
    status = initAddAddress("finish", emailCode, AutentificatorCode, envVars["OKX_AUTORIZATION"], envVars["OKX_DEVID"], addressStr)
    if status == 200 {
        log.Printf("Отправка подтверджения на биржу удалась")
        break
    }
    attempt++
    if attempt > maxAttempts {
        log.Printf("Достигнуто максимальное количество попыток. Адреса не были добавлены")
        attempt = 0
        break
    }
    log.Printf("Отправка подтверждения на биржу не удалось. Повторная попытка %d...\n", attempt)
    time.Sleep(time.Second * 10)
  }
}


func addressStrGen(wallets []string) string {
  var addressInfos []AddressInfo

	for i, address := range wallets {
		addressInfo := AddressInfo{
			Address:      address,
			ValidateName: fmt.Sprintf("addressInfoList.%d.address", i),
		}

		addressInfos = append(addressInfos, addressInfo)
	}

	jsonData, err := json.Marshal(addressInfos)
	if err != nil {
		fmt.Println("Ошибка при преобразовании в JSON:", err)
	}

	return string(jsonData)
}

func authCode(secretKey string) (code string){
	code, err := totp.GenerateCode(secretKey, time.Now())
	if err != nil {
		fmt.Println("Ошибка генерации кода аутентификации:", err)
		return
	}
	return code
}

func imapClient(imapServer string, login string, password string) string {
  timeStart := time.Now().Add(-3 * time.Hour).Add(-3 * time.Second)
	// Подключение к серверу IMAP
	c, err := client.DialTLS(imapServer, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Logout()

	// Аутентификация
	if err := c.Login(login, password); err != nil {
		log.Fatal(err)
	}
  for {
  	mbox, err := c.Select("INBOX", false)
  	if err != nil {
  		log.Println("Ошибка доступа к почтовому ящику")
  	}

  	// Получение UID последнего сообщения
  	if mbox.Messages == 0 {
  		log.Println("Почтовый ящик пуст")
  	}

    section :=  imap.BodySectionName{}
    seqSet := new(imap.SeqSet)
  	seqSet.AddNum(mbox.Messages)
  	messages := make(chan *imap.Message, 10)
  	done := make(chan error, 1)
  	go func() {
  		done <- c.Fetch(seqSet, []imap.FetchItem{section.FetchItem(), imap.FetchEnvelope}, messages)
  	}()

    // Обработка полученных сообщений
  	for msg := range messages {
      msgBody := msg.GetBody(&section)
    	if msgBody == nil {
    		log.Println("Не удалось получить текст сообщения")
    	}

      msgBodyReaded, err := mail.CreateReader(msgBody)
    	if err != nil {
    		log.Println("Не удалось прочитать сообщение")
    	}

    	// Print some info about the message
    	header := msgBodyReaded.Header

      timeMessage, err := header.Date();
    	from := msg.Envelope.From[0].Address();

  	  timeStartFormatted := timeStart.Format(timeFormat)
  	  timeMessageFormatted := timeMessage.Format(timeFormat)

      // log.Printf("Время 1: %s", timeStartFormatted)
	    // log.Printf("Время 2: %s", timeMessageFormatted)

      if timeStartFormatted <= timeMessageFormatted && strings.Contains(from, senderEmail) {
        // Process each message's part
      	for {
      		p, err := msgBodyReaded.NextPart()
      		if err == io.EOF {
      			break
      		} else if err != nil {
      		}
    			b, _ := ioutil.ReadAll(p.Body)
          re := regexp.MustCompile(`<div class="code"[^>]*>([^<]+)</div>`)
        	// Ищем совпадение в тексте сообщения
        	matches := re.FindStringSubmatch(string(b))
        	if len(matches) >= 2 {
        		value := matches[1]
            return value
        	} else {
        		fmt.Println("Собщение было найдено, но код с него не получен")
        	}
      	}
    	}
    }
    time.Sleep(time.Second * 10)
  }
}

func sendMailCode(auth string, devId string, chunk []string) int {
  // Создаем клиент HTTP
	client := &http.Client{}

	// Создаем GET-запрос
  chunkWallets := strings.Join(chunk, ",")
  url := "https://www.okx.com/v2/asset/withdraw/add/address/sendEmailCode?t=1686344378146&addressStr=" + chunkWallets + "&includeAuth=true"
	req, err := http.NewRequest("GET", url, nil)
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

func initAddAddress(stage string, emailCode string, toptCode string, auth string, devId string, addressStr string) (status int){
  url := "https://www.okx.com/v2/asset/withdraw/addressBatch?t=1686344378146"

  var payload []byte
  switch stage {
  case "start":
    payload = []byte(`{"chooseChain":true,"formGroupIndexes":[0],"authFlag":true,"subCurrencyId":2,"generalType":1,"targetType":-1,"currencyId":"2","addressInfoList":` + addressStr + `,"includeAuth":true,"whiteFlag":1,"validateOnly":true}`)
  case "finish":
    payload = []byte(`{"emailCode":"` + emailCode + `","totpCode":"` + toptCode + `","_allow":true,"chooseChain":true,"formGroupIndexes":[0],"authFlag":true,"subCurrencyId":2,"generalType":1,"targetType":-1,"currencyId":"2","addressInfoList":` + addressStr + `,"includeAuth":true,"whiteFlag":1,"validateOnly":false}`)
  }

	// Создаем новый HTTP-запрос типа POST
	req, err := http.NewRequest("POST", url, bytes.NewReader(payload))
	if err != nil {
		fmt.Println("Ошибка при создании запроса:", err)
		return
	}

	// Устанавливаем заголовки запроса
  req.Header.Add("Authorization", auth)
  req.Header.Add("Devid", devId)
  req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Создаем клиент HTTP и отправляем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Ошибка при выполнении запроса:", err)
		return
	}
	defer resp.Body.Close()

  status = resp.StatusCode
	return
}


// func sendRequest()
