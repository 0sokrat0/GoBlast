package worker

import (
	"errors"

	"GoBlast/pkg/logger"
	"go.uber.org/zap"
	tele "gopkg.in/telebot.v4"
)

// sendPhoto отправляет фото.
// Принимает *полный* TaskItem (в частности, item.TaskID можно использовать для логирования).
func (w *Worker) sendPhoto(item TaskItem) error {
	c := item.Content
	if c.MediaID == "" && c.MediaURL == "" {
		logger.Log.Warn("[Worker] sendPhoto: нет MediaID или MediaURL",
			zap.String("task_id", item.TaskID),
			zap.Int64("recipient", item.Recipient))
		return errors.New("photo: no MediaID or MediaURL")
	}

	var photo *tele.Photo
	if c.MediaID != "" {
		photo = &tele.Photo{
			File:    tele.File{FileID: c.MediaID},
			Caption: c.Caption,
		}
	} else {
		photo = &tele.Photo{
			File:    tele.FromURL(c.MediaURL),
			Caption: c.Caption,
		}
	}

	_, err := w.Bot.Send(tele.ChatID(item.Recipient), photo)
	return err
}

// sendAnimation отправляет анимацию (GIF).
func (w *Worker) sendAnimation(item TaskItem) error {
	c := item.Content
	if c.MediaID == "" && c.MediaURL == "" {
		logger.Log.Warn("[Worker] sendAnimation: нет MediaID или MediaURL",
			zap.String("task_id", item.TaskID),
			zap.Int64("recipient", item.Recipient))
		return errors.New("animation: no MediaID or MediaURL")
	}

	var anim *tele.Animation
	if c.MediaID != "" {
		anim = &tele.Animation{
			File:    tele.File{FileID: c.MediaID},
			Caption: c.Caption,
		}
	} else {
		anim = &tele.Animation{
			File:    tele.FromURL(c.MediaURL),
			Caption: c.Caption,
		}
	}

	_, err := w.Bot.Send(tele.ChatID(item.Recipient), anim)
	return err
}

// sendVideo отправляет видео.
func (w *Worker) sendVideo(item TaskItem) error {
	c := item.Content
	if c.MediaID == "" && c.MediaURL == "" {
		logger.Log.Warn("[Worker] sendVideo: нет MediaID или MediaURL",
			zap.String("task_id", item.TaskID),
			zap.Int64("recipient", item.Recipient))
		return errors.New("video: no MediaID or MediaURL")
	}

	var video *tele.Video
	if c.MediaID != "" {
		video = &tele.Video{
			File:    tele.File{FileID: c.MediaID},
			Caption: c.Caption,
		}
	} else {
		video = &tele.Video{
			File:    tele.FromURL(c.MediaURL),
			Caption: c.Caption,
		}
	}

	_, err := w.Bot.Send(tele.ChatID(item.Recipient), video)
	return err
}

// sendDocument отправляет документ (файл).
func (w *Worker) sendDocument(item TaskItem) error {
	c := item.Content
	if c.MediaID == "" && c.MediaURL == "" {
		logger.Log.Warn("[Worker] sendDocument: нет MediaID или MediaURL",
			zap.String("task_id", item.TaskID),
			zap.Int64("recipient", item.Recipient))
		return errors.New("document: no MediaID or MediaURL")
	}

	var doc *tele.Document
	if c.MediaID != "" {
		doc = &tele.Document{
			File:    tele.File{FileID: c.MediaID},
			Caption: c.Caption,
		}
	} else {
		doc = &tele.Document{
			File:    tele.FromURL(c.MediaURL),
			Caption: c.Caption,
		}
	}

	_, err := w.Bot.Send(tele.ChatID(item.Recipient), doc)
	return err
}

// sendAudio отправляет аудио.
func (w *Worker) sendAudio(item TaskItem) error {
	c := item.Content
	if c.MediaID == "" && c.MediaURL == "" {
		logger.Log.Warn("[Worker] sendAudio: нет MediaID или MediaURL",
			zap.String("task_id", item.TaskID),
			zap.Int64("recipient", item.Recipient))
		return errors.New("audio: no MediaID or MediaURL")
	}

	var audio *tele.Audio
	if c.MediaID != "" {
		audio = &tele.Audio{
			File:    tele.File{FileID: c.MediaID},
			Caption: c.Caption,
		}
	} else {
		audio = &tele.Audio{
			File:    tele.FromURL(c.MediaURL),
			Caption: c.Caption,
		}
	}

	_, err := w.Bot.Send(tele.ChatID(item.Recipient), audio)
	return err
}

// sendCircle отправляет круговое видео (VideoNote).
func (w *Worker) sendCircle(item TaskItem) error {
	c := item.Content
	if c.MediaID == "" && c.MediaURL == "" {
		logger.Log.Warn("[Worker] sendCircle: нет MediaID или MediaURL",
			zap.String("task_id", item.TaskID),
			zap.Int64("recipient", item.Recipient))
		return errors.New("circle: no MediaID or MediaURL")
	}

	var vn *tele.VideoNote
	if c.MediaID != "" {
		vn = &tele.VideoNote{
			File:     tele.File{FileID: c.MediaID},
			Length:   240, // Можно менять
			Duration: 18,  // Можно менять
		}
	} else {
		vn = &tele.VideoNote{
			File:     tele.FromURL(c.MediaURL),
			Length:   240,
			Duration: 18,
		}
	}

	_, err := w.Bot.Send(tele.ChatID(item.Recipient), vn)
	return err
}
