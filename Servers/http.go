package Servers

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
	"ywadi/crimsonq/Defs"
	"ywadi/crimsonq/Structs"
	"ywadi/crimsonq/Utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	jwtware "github.com/gofiber/jwt/v3"
	"github.com/golang-jwt/jwt/v4"
	"github.com/spf13/viper"
)

var app *fiber.App

type PostBody struct {
	ConsumerId    string `json:"consumerId" xml:"consumerId" form:"consumerId"`
	Topics        string `json:"topics" xml:"topics" form:"topics"`
	MessageId     string `json:"messageId" xml:"messageId" form:"messageId"`
	ErrMsg        string `json:"errMsg" xml:"errMsg" form:"errMsg"`
	TopicString   string `json:"topicString" xml:"topicString" form:"topicString"`
	MessageString string `json:"messageString" xml:"messageString" form:"messageString"`
	Status        string `json:"status" xml:"status" form:"status"`
	Concurrency   int    `json:"concurrency" xml:"concurrency" form:"concurrency"`
}

type PostAuthBody struct {
	Username string `json"username" xml:"username" form:"username"`
	Password string `json"password" xml:"password" form:"password"`
}

func HTTP_Start(cq *Structs.S_GOQ) {
	fmt.Println("Starting Web Server.")
	app = fiber.New()
	app.Post("/login", login)

	app.Use(recover.New())
	app.Use(cors.New())

	app.Use(func(c *fiber.Ctx) error {
		if viper.GetString("HTTP.ip_whitelist") != "*" {
			grant := Utils.SliceContains(viper.GetStringSlice("HTTP.ip_whitelist"), c.IP())
			if !grant {
				return fiber.NewError(fiber.StatusForbidden)
			}
		}
		c.Next()
		return nil
	})

	app.Static("/", "../WebUI/dist")

	app.Use("/api/", jwtware.New(jwtware.Config{
		SigningKey: []byte("crimsonQ"),
	}))

	app.Get("/checkToken", checkToken)

	for k, v := range Commands {
		if v.HTTP_Method == Defs.HTTP_GET {
			route := "/api/" + strings.ReplaceAll(k, ".", "/")
			for _, av := range v.ArgsCmd {
				route = route + "/:" + av
			}
			app.Get(route, v.HTTP_Function)
		} else if v.HTTP_Method == Defs.HTTP_POST {
			route := "/api/" + strings.ReplaceAll(k, ".", "/")
			app.Post(route, v.HTTP_Function)
		}

	}

	app.Listen(":" + viper.GetString("HTTP.port"))
}

func checkToken(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	name := claims["name"].(string)
	return c.SendString("Welcome " + name)
}

func login(c *fiber.Ctx) error {
	bodyData := PostAuthBody{}
	if err := c.BodyParser(&bodyData); err != nil {
		log.Println(err)
		return fiber.NewError(fiber.StatusBadRequest, Defs.ERRIncorrectArgs)
	}
	user := bodyData.Username
	pass := bodyData.Password

	// Throws Unauthorized error
	if user != viper.GetString("HTTP.username") || pass != viper.GetString("HTTP.password") {
		return c.SendStatus(fiber.StatusUnauthorized)
	}

	// Create the Claims
	claims := jwt.MapClaims{
		"name":  bodyData.Username,
		"admin": true,
		"exp":   time.Now().Add(time.Hour * 72).Unix(),
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Generate encoded token and send it as response.
	t, err := token.SignedString([]byte("crimsonQ"))
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(fiber.Map{"token": t})
}

func HTTP_Ping(c *fiber.Ctx) error {
	return c.JSON("Pong! " + c.Params("messageString"))
}

func HTTP_Quit(c *fiber.Ctx) error {
	return fiber.NewError(fiber.StatusMethodNotAllowed, Defs.ERRStatusNotAllowed)
}

func HTTP_Auth(c *fiber.Ctx) error {
	return fiber.NewError(fiber.StatusMethodNotAllowed, Defs.ERRStatusNotAllowed)
}

func HTTP_Command(c *fiber.Ctx) error {
	list := []string{}
	for k, v := range Commands {
		list = append(list, (k + " [" + strings.Join(v.ArgsCmd, "] [") + "]"))
	}
	return c.JSON(list)
}

//TODO: Open websocket here
func HTTP_Subscribe(c *fiber.Ctx) error {
	return fiber.NewError(fiber.StatusMethodNotAllowed, Defs.ERRStatusNotAllowed)
}

func HTTP_Exists(c *fiber.Ctx) error {
	return c.JSON(strconv.FormatBool(crimsonQ.ConsumerExists(string(c.Params("consumerId")))))
}

func HTTP_ConsumerInfo(c *fiber.Ctx) error {
	return c.JSON(crimsonQ.ConsumerInfo(c.Params("consumerId")))
}

func HTTP_Consumer_Create(c *fiber.Ctx) error {
	bodyData := PostBody{}
	if err := c.BodyParser(&bodyData); err != nil {
		log.Println(err)
		return fiber.NewError(fiber.StatusBadRequest, Defs.ERRIncorrectArgs)
	}

	consumerId := bodyData.ConsumerId
	if crimsonQ.ConsumerExists(consumerId) {
		return fiber.NewError(fiber.StatusBadRequest, Defs.ERROConsumerAlreadyExists)
	}
	consumerTopics := bodyData.Topics
	consumerConcurrent := strconv.Itoa(bodyData.Concurrency)
	crimsonQ.CreateQDB(consumerId, viper.GetString("crimson_settings.data_rootpath"))
	crimsonQ.SetConcurrency(consumerId, consumerConcurrent)
	crimsonQ.SetTopics(consumerId, consumerTopics)
	return c.JSON("Created:" + consumerId)
}

func HTTP_Set_Concurrency(c *fiber.Ctx) error {
	bodyData := PostBody{}
	if err := c.BodyParser(&bodyData); err != nil {
		fmt.Println(err)
		return fiber.NewError(fiber.StatusBadRequest, Defs.ERRIncorrectArgs)
	}
	consumerId := bodyData.ConsumerId
	concurrency := bodyData.Concurrency
	if crimsonQ.ConsumerExists(consumerId) {
		crimsonQ.SetConcurrency(consumerId, strconv.Itoa(concurrency))
		return c.JSON("ok")
	} else {
		err := errors.New(Defs.ERRincorrectConsumerId)
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprint(err))
	}
}

func HTTP_SetConsumerTopics(c *fiber.Ctx) error {
	consumerId := string(c.Params("consumerId"))
	topicFilters := string(c.Params("topics"))
	if crimsonQ.ConsumerExists(consumerId) {
		crimsonQ.SetTopics(consumerId, topicFilters)
		return c.JSON("ok")
	} else {
		err := errors.New(Defs.ERRincorrectConsumerId)
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprint(err))
	}
}

func HTTP_ConsumerConcurrencyOk(c *fiber.Ctx) error {
	consumerId := string(c.Params("consumerId"))
	if crimsonQ.ConsumerExists(consumerId) {
		return c.JSON(strconv.FormatBool(crimsonQ.ConcurrencyOk(consumerId)))
	} else {
		err := errors.New(Defs.ERRincorrectConsumerId)
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprint(err))
	}
}

func HTTP_GetConsumerTopics(c *fiber.Ctx) error {
	consumerId := string(c.Params("consumerId"))
	if crimsonQ.ConsumerExists(consumerId) {
		topics := crimsonQ.GetTopics(consumerId)
		return c.JSON(topics)
	} else {
		return fiber.NewError(fiber.StatusBadRequest, Defs.ERRincorrectConsumerId)
	}
}
func HTTP_Destroy(c *fiber.Ctx) error {
	bodyData := PostBody{}
	if err := c.BodyParser(&bodyData); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, Defs.ERRIncorrectArgs)
	}
	consumerId := bodyData.ConsumerId
	if crimsonQ.ConsumerExists(consumerId) {
		crimsonQ.DestroyQDB(consumerId)
		return c.SendString("ok")
	} else {
		return fiber.NewError(fiber.StatusBadRequest, Defs.ERRincorrectConsumerId)
	}
}

func HTTP_List(c *fiber.Ctx) error {
	return c.JSON(crimsonQ.ListConsumers())
}

func HTTP_Msg_Keys(c *fiber.Ctx) error {
	consumerId := string(c.Params("consumerId"))
	list, err := crimsonQ.ListAllKeys(consumerId)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprint(err))

	}
	resp := []string{}
	resp = append(resp, list...)
	return c.JSON(resp)
}

func HTTP_Msg_Counts(c *fiber.Ctx) error {
	consumerId := string(c.Params("consumerId"))
	counts, err := crimsonQ.GetKeyCount(consumerId)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, fmt.Sprint(err))
	}
	return c.JSON(counts)
}

func HTTP_Msg_Push_Topic(c *fiber.Ctx) error {
	bodyData := PostBody{}
	if err := c.BodyParser(&bodyData); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, Defs.ERRIncorrectArgs)
	}
	topic := bodyData.TopicString
	message := bodyData.MessageString
	consumers := crimsonQ.PushTopic(topic, message)
	for consumer, msgkey := range consumers {
		msgkeySplit := strings.Split(msgkey, ":")
		PS.Publish(consumer, msgkeySplit[1])
	}
	return c.JSON("Ok")
}

func HTTP_Msg_Push_Consumer(c *fiber.Ctx) error {
	bodyData := PostBody{}
	if err := c.BodyParser(&bodyData); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, Defs.ERRIncorrectArgs)
	}
	consumerId := bodyData.ConsumerId
	topic := bodyData.ConsumerId
	message := bodyData.MessageString
	if crimsonQ.ConsumerExists(consumerId) {
		msgkey := crimsonQ.PushConsumer(consumerId, "direct:"+topic, message)
		msgkeySplit := strings.Split(msgkey, ":")
		PS.Publish(consumerId, msgkeySplit[1])
		return c.JSON("Ok")
	} else {
		return fiber.NewError(fiber.StatusBadRequest, Defs.ERRincorrectConsumerId)
	}

}

func HTTP_Msg_Pull(c *fiber.Ctx) error {
	consumerId := string(c.Params("consumerId"))
	if crimsonQ.ConsumerExists(consumerId) {
		msg, err := crimsonQ.Pull(consumerId)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())

		}
		return c.JSON(msg)
	} else {
		return fiber.NewError(fiber.StatusBadRequest, Defs.ERRincorrectConsumerId)
	}
}

func HTTP_Msg_Del(c *fiber.Ctx) error {
	bodyData := PostBody{}
	if err := c.BodyParser(&bodyData); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, Defs.ERRIncorrectArgs)
	}

	consumerId := bodyData.ConsumerId
	status := bodyData.Status
	messageId := bodyData.MessageId

	if crimsonQ.ConsumerExists(consumerId) {
		err := crimsonQ.Del(status, consumerId, messageId)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprint(err))
		}
		return c.JSON(("Ok"))
	} else {
		return fiber.NewError(fiber.StatusBadRequest, Defs.ERRincorrectConsumerId)
	}
}

func HTTP_Msg_Get_Status_Json(c *fiber.Ctx) error {
	consumerId := string(c.Params("consumerId"))
	status := strings.ToLower(c.Params("status"))
	if !(status == Defs.STATUS_ACTIVE || status == Defs.STATUS_COMPLETED || status == Defs.STATUS_DELAYED || status == Defs.STATUS_FAILED || status == Defs.STATUS_PENDING) {
		sError := errors.New(Defs.ERRIncorrectStatus)
		return fiber.NewError(fiber.StatusBadRequest, sError.Error())
	}
	if crimsonQ.ConsumerExists(consumerId) {
		json, err := crimsonQ.GetAllByStatusJson(consumerId, status)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprint(err))
		}
		c.JSON(json)
	} else {
		return fiber.NewError(fiber.StatusBadRequest, Defs.ERRincorrectConsumerId)
	}
	return nil
}

func HTTP_Msg_Fail(c *fiber.Ctx) error {
	bodyData := PostBody{}
	if err := c.BodyParser(&bodyData); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, Defs.ERRIncorrectArgs)
	}

	consumerId := bodyData.ConsumerId
	messageId := bodyData.MessageId
	errMsg := bodyData.ErrMsg
	if crimsonQ.ConsumerExists(consumerId) {
		err := crimsonQ.MsgFail(consumerId, messageId, errMsg)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, Defs.ERRIncorrectMessageId)
		}
		return c.SendString("Ok")
	} else {
		return fiber.NewError(fiber.StatusBadRequest, Defs.ERRincorrectConsumerId)
	}
}

func HTTP_Msg_Complete(c *fiber.Ctx) error {
	bodyData := PostBody{}
	if err := c.BodyParser(&bodyData); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, Defs.ERRIncorrectArgs)
	}

	consumerId := bodyData.ConsumerId
	messageId := bodyData.MessageId
	if crimsonQ.ConsumerExists(consumerId) {
		err := crimsonQ.MsgComplete(consumerId, messageId)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, fmt.Sprint(err))
		}
		return c.SendString("Ok")
	} else {
		return fiber.NewError(fiber.StatusBadRequest, Defs.ERRincorrectConsumerId)
	}
}

func HTTP_Msg_Retry(c *fiber.Ctx) error {
	bodyData := PostBody{}
	if err := c.BodyParser(&bodyData); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, Defs.ERRIncorrectArgs)
	}
	consumerId := bodyData.ConsumerId
	messageId := bodyData.MessageId
	if crimsonQ.ConsumerExists(consumerId) {
		err := crimsonQ.MsgRetry(consumerId, messageId)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, Defs.ERRincorrectConsumerId)
		}
		return c.SendString("Ok")
	} else {
		return fiber.NewError(fiber.StatusBadRequest, Defs.ERRincorrectConsumerId)
	}
}

func HTTP_Msg_Retry_All(c *fiber.Ctx) error {
	bodyData := PostBody{}
	if err := c.BodyParser(&bodyData); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, Defs.ERRIncorrectArgs)
	}
	consumerId := bodyData.ConsumerId
	if crimsonQ.ConsumerExists(consumerId) {
		crimsonQ.ReqAllFailed(consumerId)
		return c.SendString("Ok")
	} else {
		return fiber.NewError(fiber.StatusBadRequest, Defs.ERRincorrectConsumerId)
	}
}

func HTTP_Flush_Complete(c *fiber.Ctx) error {
	bodyData := PostBody{}
	if err := c.BodyParser(&bodyData); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, Defs.ERRIncorrectArgs)
	}
	consumerId := bodyData.ConsumerId
	if crimsonQ.ConsumerExists(consumerId) {
		crimsonQ.ClearComplete(consumerId)
		return c.SendString("Ok")
	} else {
		return fiber.NewError(fiber.StatusBadRequest, Defs.ERRincorrectConsumerId)
	}
}

func HTTP_Flush_Failed(c *fiber.Ctx) error {
	bodyData := PostBody{}
	if err := c.BodyParser(&bodyData); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, Defs.ERRIncorrectArgs)
	}
	consumerId := bodyData.ConsumerId
	if crimsonQ.ConsumerExists(consumerId) {
		crimsonQ.ClearFailed(consumerId)
		return c.SendString("Ok")

	} else {
		return fiber.NewError(fiber.StatusBadRequest, Defs.ERRincorrectConsumerId)
	}
}

func HTTP_Info(c *fiber.Ctx) error {
	info := []string{"CrimsonQ Server", "A NextGen Message Queue!"}
	return c.JSON(info)
}
