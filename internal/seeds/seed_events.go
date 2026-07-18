package seeds

import (
	"fmt"
	"log"

	"pleco-api/internal/modules/event"

	"gorm.io/gorm"
)

func SeedEvents(db *gorm.DB) {
	mustHaveDB(db)

	events := []event.Event{
		{
			ExternalID:   "bakpia-trail-festival",
			Title:        "Bakpia Trail Festival",
			Description:  "Bakpia Trail Festival adalah perayaan kuliner dan budaya yang menyoroti kelezatan bakpia, camilan legendaris Yogyakarta. Nikmati lokakarya memasak, sesi mencicipi, pasar makanan, dan pertunjukan budaya yang menampilkan tradisi lokal.\n\nAcara ini menggabungkan cita rasa, kreativitas, dan warisan dalam satu pengalaman yang tak terlupakan.",
			Location:     "Pathuk, Yogyakarta",
			StartDate:    "2026-08-15",
			EndDate:      "2026-08-16",
			ImageURL:     "https://images.unsplash.com/photo-1555126634-323283e090fa?auto=format&fit=crop&w=1600&q=80",
			Category:     "Food Festival",
			Status:       "active",
			Latitude:     -7.8024,
			Longitude:    110.3649,
			MaxAttendees: 2000,
			TicketPrice:  "IDR 100,000",
			Organizer:    "Yogyakarta Culinary Board",
			Highlights:   event.JSONArr{"Bakpia Cooking Workshop", "Street Food Market", "Cultural Performances"},
		},
		{
			ExternalID:   "sekaten-2026",
			Title:        "Sekaten Yogyakarta 2026",
			Description:  "Sekaten adalah festival budaya tahunan yang merayakan kelahiran Nabi Muhammad SAW. Festival ini menampilkan gamelan kerajaan, pasar malam, dan berbagai pertunjukan seni tradisional Jawa di sekitar Alun-Alun Utara Keraton Yogyakarta.",
			Location:     "Alun-Alun Utara, Yogyakarta",
			StartDate:    "2026-09-01",
			EndDate:      "2026-09-07",
			ImageURL:     "https://images.unsplash.com/photo-1533050487297-09b450131914?auto=format&fit=crop&w=1600&q=80",
			Category:     "Cultural Festival",
			Status:       "upcoming",
			Latitude:     -7.8052,
			Longitude:    110.3647,
			MaxAttendees: 10000,
			TicketPrice:  "Gratis",
			Organizer:    "Keraton Yogyakarta",
			Highlights:   event.JSONArr{"Gamelan Kerajaan", "Pasar Malam Sekaten", "Gunungan Grebeg", "Pertunjukan Wayang"},
		},
		{
			ExternalID:   "prambanan-jazz-2026",
			Title:        "Prambanan Jazz Festival 2026",
			Description:  "Prambanan Jazz Festival adalah salah satu festival musik paling ikonik di Indonesia. Digelar di depan latar belakang megah Candi Prambanan yang diterangi cahaya, festival ini menghadirkan musisi jazz terbaik dari Indonesia dan dunia.",
			Location:     "Kompleks Candi Prambanan, Sleman",
			StartDate:    "2026-07-18",
			EndDate:      "2026-07-20",
			ImageURL:     "https://images.unsplash.com/photo-1470229722913-7c0e2dbbafd3?auto=format&fit=crop&w=1600&q=80",
			Category:     "Music Festival",
			Status:       "active",
			Latitude:     -7.7520,
			Longitude:    110.4914,
			MaxAttendees: 5000,
			TicketPrice:  "IDR 350,000",
			Organizer:    "Rajawali Indonesia",
			Highlights:   event.JSONArr{"Live Jazz Performance", "Sunset Concert", "Heritage Stage", "Culinary Village"},
		},
		{
			ExternalID:   "merapi-adventure-race",
			Title:        "Merapi Adventure Race 2026",
			Description:  "Tantang dirimu dalam kompetisi petualangan multi-disiplin melintasi medan vulkanik Gunung Merapi. Termasuk trail running, mountain biking, dan navigasi rute. Cocok untuk petualang sejati yang ingin merasakan thrill Merapi.",
			Location:     "Lereng Merapi, Sleman",
			StartDate:    "2026-08-22",
			EndDate:      "2026-08-23",
			ImageURL:     "https://images.unsplash.com/photo-1556375403-b96342fc0ee2?auto=format&fit=crop&w=1600&q=80",
			Category:     "Adventure",
			Status:       "active",
			Latitude:     -7.5407,
			Longitude:    110.4457,
			MaxAttendees: 500,
			TicketPrice:  "IDR 250,000",
			Organizer:    "Merapi Volcano Trail Community",
			Highlights:   event.JSONArr{"Trail Running 42km", "Mountain Biking", "Volcanic Landscape Views", "Night Navigation"},
		},
		{
			ExternalID:   "jogja-fashion-week",
			Title:        "Jogja Fashion Week 2026",
			Description:  "Pekan mode bergengsi yang menampilkan koleksi terbaru dari desainer lokal dan nasional dengan sentuhan batik dan tenun tradisional Yogyakarta. Merayakan kekayaan tekstil Jawa dalam balutan mode kontemporer.",
			Location:     "Jogja Expo Center, Yogyakarta",
			StartDate:    "2026-10-10",
			EndDate:      "2026-10-13",
			ImageURL:     "https://images.unsplash.com/photo-1558769132-cb1aea458c5e?auto=format&fit=crop&w=1600&q=80",
			Category:     "Fashion & Art",
			Status:       "upcoming",
			Latitude:     -7.8226,
			Longitude:    110.4107,
			MaxAttendees: 3000,
			TicketPrice:  "IDR 75,000",
			Organizer:    "Dinas Pariwisata DIY",
			Highlights:   event.JSONArr{"Batik Runway Show", "Traditional Weaving Exhibition", "Designer Meet & Greet", "Craft Market"},
		},
		{
			ExternalID:   "parangtritis-kite-festival",
			Title:        "Parangtritis International Kite Festival",
			Description:  "Festival layang-layang internasional yang digelar di pantai ikonik Parangtritis. Ratusan layang-layang raksasa dari berbagai negara menghiasi langit pantai selatan Yogyakarta. Pertunjukan terbaik saat angin laut berhembus kencang.",
			Location:     "Pantai Parangtritis, Bantul",
			StartDate:    "2026-09-19",
			EndDate:      "2026-09-21",
			ImageURL:     "https://images.unsplash.com/photo-1530870110042-98b2cb110834?auto=format&fit=crop&w=1600&q=80",
			Category:     "Cultural Festival",
			Status:       "upcoming",
			Latitude:     -8.0257,
			Longitude:    110.3318,
			MaxAttendees: 8000,
			TicketPrice:  "Gratis",
			Organizer:    "Dinas Pariwisata Bantul",
			Highlights:   event.JSONArr{"International Kite Competition", "Night Kite Show", "Beach Sunset View", "Cultural Performance"},
		},
	}

	for _, e := range events {
		existing := event.Event{}
		if err := db.Where("external_id = ?", e.ExternalID).First(&existing).Error; err == nil {
			continue // already seeded
		}
		if err := db.Create(&e).Error; err != nil {
			log.Printf("Failed to seed event %s: %v", e.ExternalID, err)
			continue
		}
	}
	fmt.Println("Events seeding done")
}
