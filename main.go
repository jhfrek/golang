package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand/v2"
	"os"
	"strconv"
	"time"

	_ "github.com/lib/pq"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "Simbirka"
	dbname   = "telegram_users_db"
)

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}

func getEnvVariable(key string) string {
	// load key.env file
	err := godotenv.Load("key.env")
	if err != nil {
		log.Fatalf("Error loading key.env file")
	}
	return os.Getenv(key)
}

// Генератор случайных чисел
func randRange(min, max int) int {
	return rand.IntN(max-min) + min
}

type UserData struct {
	user_id      int64
	data_time    string
	registration bool
}

func main() {
	bot, err := tgbotapi.NewBotAPI(getEnvVariable("TG_API_KEY"))
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Setup long-polling request
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	psqlconn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open("postgres", psqlconn)
	CheckError(err)
	defer db.Close()

	//_, err = db.Exec(`CREATE TABLE telegram_users_tab(User_ID INTEGER PRIMARY KEY, Data_Time TIMESTAMP NOT NULL, Registration BOOLEAN);`)
	//CheckError(err)

	defer db.Close()
	err = db.Ping()
	CheckError(err)

	numAnswer := ""
	numAnswerP := &numAnswer
	survey := false
	surveyP := &survey
	boolQuery := false
	boolQueryP := &boolQuery
	var glUserId int64 = 0
	glUserIdP := &glUserId
	oldUser := false
	oldUserP := &oldUser

	// Выведем горутину в отдельный поток для опроса просроченных данных
	go func() {
		for {
			time.Sleep(time.Minute)
			if *glUserIdP > 0 {
				if *glUserIdP == 0 {
					break
				}
				strSelect := fmt.Sprintf("SELECT * FROM telegram_users_tab WHERE user_id = '%d' AND NOT registration AND Data_Time < '%s';", *glUserIdP, time.Now().Add(time.Duration(-1) * time.Hour).String()[:19])
				dbRes, err := db.Query(strSelect)
				CheckError(err)
				*oldUserP = false
				for dbRes.Next() {
					*oldUserP = true
					break
				}
				dbRes.Close()
				if *oldUserP {
					delData := `DELETE FROM telegram_users_tab WHERE user_id = $1`
					_, err := db.Exec(delData, *glUserIdP)
					CheckError(err)
				}
			}
		}
	}()
	// Обрабатываем сообщения из чата
	for update := range updates {
		if update.Message != nil { // Есть новое сообщение
			text := update.Message.Text      // Текст сообщения
			userID := update.Message.From.ID // ID пользователя
			*glUserIdP = userID
			var replyMsg string

			log.Printf("[%s](%d) %s", update.Message.From.UserName, userID, text)

			// Анализируем текст сообщения и записываем ответ в переменную

			*boolQueryP = false
			strSelect := fmt.Sprintf("SELECT * FROM telegram_users_tab WHERE user_id = '%d';", userID)
			dbRes, err := db.Query(strSelect)
			CheckError(err)
			isRecord := false
			isRecordP := &isRecord
			userDataSl := make([]*UserData, 0)
			for dbRes.Next() {
				*isRecordP = true
				userDataStruc := new(UserData)
				err := dbRes.Scan(&userDataStruc.user_id, &userDataStruc.data_time, &userDataStruc.registration)
				CheckError(err)
				userDataSl = append(userDataSl, userDataStruc)
			}
			dbRes.Close()
			if *isRecordP {
				for _, userDa := range userDataSl {
					if userDa.registration {
						*glUserIdP = 0
					} else {
						if *surveyP {
							if *numAnswerP == text {
								insertData := `UPDATE telegram_users_tab SET registration = true WHERE user_id = $1`
								_, err := db.Exec(insertData, userID)
								CheckError(err)
								replyMsg = "Верно! Вы зарегистрированы в системе!"
								msg := tgbotapi.NewMessage(userID, replyMsg)
								bot.Send(msg)
							} else {
								*boolQueryP = true
								replyMsg = "Ошибка! Вы неверно ввели значение выражения!"
								msg := tgbotapi.NewMessage(userID, replyMsg)
								bot.Send(msg)
							}
						} else {
							*boolQueryP = true
						}
					}
				}
			}
			if *boolQueryP || !*isRecordP {
				num1 := randRange(1, 9)
				num2 := randRange(1, 9)
				*numAnswerP = strconv.Itoa(num1 + num2)
				if !*isRecordP {
					timeZ := time.Now().String()
					insertData := `insert into "telegram_users_tab"("user_id", "data_time", "registration") values($1, $2, $3)`
					_, err := db.Exec(insertData, userID, timeZ[:19], false)
					CheckError(err)
				}
				replyMsg = "Введите значение выражения " + strconv.Itoa(num1) + " + " + strconv.Itoa(num2) + " = ?"
				msg := tgbotapi.NewMessage(userID, replyMsg) // Создаем новое сообщение
				bot.Send(msg)                                // Отвечаем приватным сообщением
				*surveyP = true
			}
		}
	}
}
