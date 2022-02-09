package Servers

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"ywadi/crimsonq/Defs"
	"ywadi/crimsonq/Structs"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/spf13/viper"
)

var app *fiber.App

type PostBody struct {
	consumerId    string
	topics        string
	messageId     string
	errMsg        string
	topicString   string
	messageString string
	status        string
	concurrency   string
}

func HTTP_Start(cq *Structs.S_GOQ) {
	app = fiber.New()
	app.Use(recover.New())

	for k, v := range Commands {
		if v.HTTP_Method == Defs.HTTP_GET {
			route := "/api/" + strings.ReplaceAll(k, ".", "/")
			for _, av := range v.ArgsCmd {
				route = route + "/:" + av
			}
			fmt.Println("|GET|" + route + "|" + strings.Join(v.ArgsCmd, " - ") + "|JSON|")
			app.Get(route, v.HTTP_Function)
		} else if v.HTTP_Method == Defs.HTTP_POST {
			route := "/api/" + strings.ReplaceAll(k, ".", "/")
			fmt.Println("|POST|" + route + "|" + strings.Join(v.ArgsCmd, " - ") + "|JSON|")
		}

	}

	app.Listen(":8080")
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

func HTTP_Consumer_Create(c *fiber.Ctx) error {
	bodyData := PostBody{}
	if err := c.BodyParser(bodyData); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, Defs.ERRIncorrectArgs)
	}

	consumerId := bodyData.consumerId
	if crimsonQ.ConsumerExists(consumerId) {
		return fiber.NewError(fiber.StatusBadRequest, Defs.ERROConsumerAlreadyExists)
	}
	consumerTopics := bodyData.topics
	consumerConcurrent := bodyData.concurrency
	crimsonQ.CreateQDB(consumerId, viper.GetString("crimson_settings.data_rootpath"))
	crimsonQ.SetConcurrency(consumerId, consumerConcurrent)
	crimsonQ.SetTopics(consumerId, consumerTopics)
	return c.JSON("Created:" + consumerId)
}

func HTTP_Set_Concurrency(c *fiber.Ctx) error {
	bodyData := PostBody{}
	if err := c.BodyParser(bodyData); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, Defs.ERRIncorrectArgs)
	}
	consumerId := bodyData.consumerId
	concurrency := bodyData.concurrency
	if crimsonQ.ConsumerExists(consumerId) {
		crimsonQ.SetConcurrency(consumerId, concurrency)
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
	if err := c.BodyParser(bodyData); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, Defs.ERRIncorrectArgs)
	}
	consumerId := bodyData.consumerId
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
	for _, s := range list {
		resp = append(resp, s)
	}
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
	if err := c.BodyParser(bodyData); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, Defs.ERRIncorrectArgs)
	}
	topic := bodyData.topicString
	message := bodyData.messageString
	consumers := crimsonQ.PushTopic(topic, message)
	for consumer, msgkey := range consumers {
		msgkeySplit := strings.Split(msgkey, ":")
		PS.Publish(consumer, msgkeySplit[1])
	}
	return c.JSON("Ok")
}

func HTTP_Msg_Push_Consumer(c *fiber.Ctx) error {
	bodyData := PostBody{}
	if err := c.BodyParser(bodyData); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, Defs.ERRIncorrectArgs)
	}
	consumerId := bodyData.consumerId
	topic := bodyData.consumerId
	message := bodyData.messageString
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
	if err := c.BodyParser(bodyData); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, Defs.ERRIncorrectArgs)
	}

	consumerId := bodyData.consumerId
	status := bodyData.status
	messageId := bodyData.messageId

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
	if err := c.BodyParser(bodyData); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, Defs.ERRIncorrectArgs)
	}

	consumerId := bodyData.consumerId
	messageId := bodyData.messageId
	errMsg := bodyData.errMsg
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
	if err := c.BodyParser(bodyData); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, Defs.ERRIncorrectArgs)
	}

	consumerId := bodyData.consumerId
	messageId := bodyData.messageId
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
	if err := c.BodyParser(bodyData); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, Defs.ERRIncorrectArgs)
	}
	consumerId := bodyData.consumerId
	messageId := bodyData.messageId
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
	if err := c.BodyParser(bodyData); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, Defs.ERRIncorrectArgs)
	}
	consumerId := bodyData.consumerId
	if crimsonQ.ConsumerExists(consumerId) {
		crimsonQ.ReqAllFailed(consumerId)
		return c.SendString("Ok")
	} else {
		return fiber.NewError(fiber.StatusBadRequest, Defs.ERRincorrectConsumerId)
	}
}

func HTTP_Flush_Complete(c *fiber.Ctx) error {
	bodyData := PostBody{}
	if err := c.BodyParser(bodyData); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, Defs.ERRIncorrectArgs)
	}
	consumerId := bodyData.consumerId
	if crimsonQ.ConsumerExists(consumerId) {
		crimsonQ.ClearComplete(consumerId)
		return c.SendString("Ok")
	} else {
		return fiber.NewError(fiber.StatusBadRequest, Defs.ERRincorrectConsumerId)
	}
}

func HTTP_Flush_Failed(c *fiber.Ctx) error {
	bodyData := PostBody{}
	if err := c.BodyParser(bodyData); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, Defs.ERRIncorrectArgs)
	}
	consumerId := bodyData.consumerId
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
