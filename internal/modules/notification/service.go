package notification

type Service struct {
	Repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{Repo: repo}
}

func (s *Service) Notify(userID uint, title, content, nType string) error {
	notif := Notification{
		UserID:  userID,
		Title:   title,
		Content: content,
		Type:    nType,
		IsRead:  false,
	}
	return s.Repo.Create(&notif)
}

func (s *Service) GetUserNotifications(userID uint) ([]Notification, error) {
	return s.Repo.FindByUserID(userID)
}

func (s *Service) MarkAsRead(id uint, userID uint) error {
	return s.Repo.MarkAsRead(id, userID)
}

func (s *Service) GetUnreadCount(userID uint) (int64, error) {
	return s.Repo.CountUnread(userID)
}
