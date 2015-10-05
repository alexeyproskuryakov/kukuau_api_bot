package shop

import (
	"bytes"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"

	d "msngr/db"
	s "msngr/structs"
	"errors"
	"gopkg.in/mgo.v2"
)

type ShopConfig struct {
	Key  string `json:"key"`
	Name string `json:"name"`
	Info string `json:"information"`
}


func FormShopCommands(db *d.DbHandlerMixin, config *ShopConfig) *s.BotContext {
	var ShopRequestCommands = map[string]s.RequestCommandProcessor{
		"commands": ShopCommandsProcessor{DbHandlerMixin: *db},
	}

	var ShopMessageCommands = map[string]s.MessageCommandProcessor{
		"information":     ShopInformationProcessor{Info:config.Info},
		"authorise":       ShopAuthoriseProcessor{DbHandlerMixin: *db},
		"log_out":         ShopLogOutMessageProcessor{DbHandlerMixin: *db},
		"orders_state":    ShopOrderStateProcessor{DbHandlerMixin: *db},
		"support_message": ShopSupportMessageProcessor{},
		"balance":         ShopBalanceProcessor{},
	}

	context := s.BotContext{}
	context.Check = func() (string, bool) { return "", true }
	context.Message_commands = ShopMessageCommands
	context.Request_commands = ShopRequestCommands
	return &context
}

var authorised_commands = []s.OutCommand{
	s.OutCommand{
		Title:    "Мои заказы",
		Action:   "orders_state",
		Position: 0,
	},
	s.OutCommand{
		Title:    "Мой баланс",
		Action:   "balance",
		Position: 1,
	},
	s.OutCommand{
		Title:    "Задать вопрос",
		Action:   "support_message",
		Position: 2,
		Fixed:    true,
		Form: &s.OutForm{
			Type: "form",
			Text: "?(text)",
			Fields: []s.OutField{
				s.OutField{
					Name: "text",
					Type: "text",
					Attributes: s.FieldAttribute{
						Label:    "Текст вопроса",
						Required: true,
					},
				},
			},
		},
	},
	s.OutCommand{
		Title:    "Выйти",
		Action:   "log_out",
		Position: 3,
	},
}
var not_authorised_commands = []s.OutCommand{
	s.OutCommand{
		Title:    "Авторизоваться",
		Action:   "authorise",
		Position: 0,
		Form: &s.OutForm{
			Name: "Форма ввода данных пользователя",
			Type: "form",
			Text: "Пользователь: ?(username), пароль: ?(password)",
			Fields: []s.OutField{
				s.OutField{
					Name: "username",
					Type: "text",
					Attributes: s.FieldAttribute{
						Label:    "имя",
						Required: true,
					},
				},
				s.OutField{
					Name: "password",
					Type: "password",
					Attributes: s.FieldAttribute{
						Label:    "пароль",
						Required: true,
					},
				},
			},
		},
	},
}

func _get_user_and_password(fields []s.InField) (*string, *string) {
	var user, password *string
	for _, field := range fields {
		if field.Name == "username" {
			user = &(field.Data.Value)
		} else if field.Name == "password" {
			password = &(field.Data.Value)
		}
	}
	return user, password
}

type ShopCommandsProcessor struct {
	d.DbHandlerMixin
}

func (cp ShopCommandsProcessor) ProcessRequest(in *s.InPkg) *s.RequestResult {
	user_state, err := cp.Users.GetUserState(in.From)
	if err == mgo.ErrNotFound {
		user_data := in.UserData
		if user_data == nil {
			return s.ExceptionRequestResult(errors.New("not user data !"), &not_authorised_commands)
		}
		phone := in.UserData.Phone
		if phone == "" {
			return s.ExceptionRequestResult(errors.New("not user data phone!"), &not_authorised_commands)
		}
		cp.Users.AddUser(&(in.From), &phone)
	} else {
		return s.ExceptionRequestResult(err, &not_authorised_commands)
	}
	commands := []s.OutCommand{}
	if *user_state == d.LOGIN {
		commands = authorised_commands
	} else {
		commands = not_authorised_commands
	}
	return &s.RequestResult{Commands:&commands}
}

type ShopAuthoriseProcessor struct {
	d.DbHandlerMixin
}

func (sap ShopAuthoriseProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	command := *in.Message.Commands
	user, password := _get_user_and_password(command[0].Form.Fields)
	if user == nil || password == nil {
		return s.ExceptionMessageResult(errors.New("Не могу извлечь логин и (или) пароль."))
	}

	check, err := sap.Users.CheckUserPassword(user, password)
	if err != nil && err != mgo.ErrNotFound {
		return s.ExceptionMessageResult(err)
	}

	var body string
	var commands []s.OutCommand

	if *check {
		sap.Users.SetUserState(&(in.From), d.LOGIN)
		body = "Добро пожаловать в интернет магазин Desprice Markt!"
		commands = authorised_commands
	}else {
		body = "Не правильные логин или пароль :("
		commands = not_authorised_commands
	}
	return &s.MessageResult{Body:body, Commands:&commands}

}

type ShopOrderStateProcessor struct {
	d.DbHandlerMixin
}

func __choiceString(choices []string) string {
	var winner string
	length := len(choices)
	rand.Seed(time.Now().Unix())
	i := rand.Intn(length)
	winner = choices[i]
	return winner
}

var order_states = [5]string{"обработан", "доставляется", "отправлен", "поступил в пункт выдачи", "в обработке"}
var order_products = [4]string{"Ноутбук Apple MacBook Air", "Электрочайник BORK K 515", "Аудиосистема Westlake Tower SM-1", "Микроволновая печь Bosch HMT85ML23"}

func (osp ShopOrderStateProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	user_state, err := osp.Users.GetUserState(in.From)
	if err != nil && err != mgo.ErrNotFound {
		return s.ExceptionMessageResult(err)
	}

	var result string
	var commands []s.OutCommand
	if *user_state == d.LOGIN {
		result = fmt.Sprintf("Ваш заказ #%v (%v) %v.", rand.Int31n(10000), __choiceString(order_products[:]), __choiceString(order_states[:]))
		commands = authorised_commands
	} else {
		result = "Авторизуйтесь пожалуйста!"
		commands = not_authorised_commands
	}
	return &s.MessageResult{Body:result, Commands:&commands}
}

type ShopSupportMessageProcessor struct{}

func contains(container string, elements []string) bool {
	container_elements := regexp.MustCompile("[a-zA-Zа-яА-Я]+").FindAllString(container, -1)
	ce_map := make(map[string]bool)
	for _, ce_element := range container_elements {
		ce_map[strings.ToLower(ce_element)] = true
	}
	result := true
	for _, element := range elements {
		_, ok := ce_map[element]
		result = result && ok
	}
	return result
}

func make_one_string(fields []s.InField) string {
	var buffer bytes.Buffer
	for _, field := range fields {
		buffer.WriteString(field.Data.Value)
		buffer.WriteString(field.Data.Text)
	}
	return buffer.String()
}

func (sm ShopSupportMessageProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	commands := *in.Message.Commands
	var body string

	if commands != nil {
		if contains(make_one_string(commands[0].Form.Fields), []string{"где", "забрать", "заказ"}) {
			body = "Ваш заказ вы можете забрать по адресу: ул. Николаева д. 11."
		} else {
			body = "Спасибо за вопрос. Мы ответим Вам в ближайшее время."
		}
	} else {
		body = "Спасибо за вопрос. Мы ответим Вам в ближайшее время."
	}
	return &s.MessageResult{Body:body}
}

type ShopInformationProcessor struct{
	Info string
}

func (ih ShopInformationProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	info := ih.Info
	if info == ""{
		info = "Desprice Markt - интернет-магазин бытовой техники и электроники в Новосибирске и других городах России. Каталог товаров мировых брендов."
	}
	return &s.MessageResult{Body:info}
}

type ShopLogOutMessageProcessor struct {
	d.DbHandlerMixin
}

func (lop ShopLogOutMessageProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	err := lop.Users.SetUserState(&(in.From), d.LOGOUT)
	if err != nil {
		return s.ExceptionMessageResult(err)
	}
	return &s.MessageResult{Body:"До свидания! ", Commands:&not_authorised_commands}
}

type ShopBalanceProcessor struct {
}

func (sbp ShopBalanceProcessor) ProcessMessage(in *s.InPkg) *s.MessageResult {
	return &s.MessageResult{Body: fmt.Sprintf("Ваш баланс на %v составляет %v бонусных баллов.", time.Now().Format("01.02.2006"), rand.Int31n(1000) + 10), Commands: &authorised_commands}
}
