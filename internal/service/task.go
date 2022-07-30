package service

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha512"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/streadway/amqp"
	"github.com/teneta-io/dcp/internal/entities"
	"github.com/teneta-io/dcp/pkg/multipass"
	"github.com/teneta-io/dcp/pkg/rabbitmq"
	"go.uber.org/zap"
	"log"
)

var (
	ErrCanNotUnmarshalTask = errors.New("can not unmarshal queue task")
)

type TaskService struct {
	logger         *zap.Logger
	taskSubscriber *rabbitmq.TaskSubscriber
}

func NewTaskService(logger *zap.Logger, taskSubscriber *rabbitmq.TaskSubscriber) *TaskService {
	s := &TaskService{
		logger:         logger,
		taskSubscriber: taskSubscriber,
	}

	go func() {
		if err := s.taskSubscriber.Subscribe(s.handle); err != nil {
			s.logger.Error("Can't subscribe to queue", zap.Error(err))
		}
	}()

	return s
}

func (s *TaskService) handle(delivery amqp.Delivery) {
	task := &entities.Task{}

	if err := json.Unmarshal(delivery.Body, &task); err != nil {
		s.logger.Error(ErrCanNotUnmarshalTask.Error(), zap.Error(err))
		return
	}

	if err := s.verify(task.Payload, task.DCCPublicKey, task.DCCSign); err != nil {
		zap.S().Error(err)
		return
	}

	s.proceed(task)
}

func (s *TaskService) proceed(task *entities.Task) {
	s.logger.Info("task", zap.Any("task", task))
	instance, err := multipass.Launch(&multipass.LaunchReq{
		CPU:           int(task.Payload.Requirements.VCPU),
		Disk:          fmt.Sprintf("%dG", task.Payload.Requirements.Storage),
		Memory:        fmt.Sprintf("%dG", task.Payload.Requirements.RAM),
		CloudInitFile: "",
	})

	if err != nil {
		s.logger.Error(err.Error())
		return
	}

	log.Println(instance)
}

func (s *TaskService) verify(payload *entities.TaskPayload, publicKey *rsa.PublicKey, signature []byte) error {
	bts, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	hash := sha512.Sum512(bts)

	return rsa.VerifyPKCS1v15(publicKey, crypto.SHA512, hash[:], signature)
}
