package seeds

import (
	"fmt"
	"log"
	"time"

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
			ImageURL:     "https://commons.wikimedia.org/wiki/Special:FilePath/Bakpia%20pathuk.jpg",
			Category:     "Food Festival",
			Status:       "",
			Latitude:     -7.8024,
			Longitude:    110.3649,
			MaxAttendees: 2000,
			TicketPrice:  "IDR 100,000",
			Organizer:    "Yogyakarta Culinary Board",
			Highlights:   event.JSONArr{"Bakpia Cooking Workshop", "Street Food Market", "Cultural Performances"},
		},
		{
			ExternalID:    "sekaten-2026",
			DestinationID: "keraton",
			Title:         "Sekaten Yogyakarta 2026",
			Description:  "Sekaten adalah festival budaya tahunan yang merayakan kelahiran Nabi Muhammad SAW. Festival ini menampilkan gamelan kerajaan, pasar malam, dan berbagai pertunjukan seni tradisional Jawa di sekitar Alun-Alun Utara Keraton Yogyakarta.",
			Location:     "Alun-Alun Utara, Yogyakarta",
			StartDate:    "2026-09-01",
			EndDate:      "2026-09-07",
			ImageURL:     "https://upload.wikimedia.org/wikipedia/commons/c/c9/Gunungan_darat_during_Garebeg_Mulud_Yogyakarta_Dec_2017_Pj_IMG_4517sm.jpg",
			Category:     "Cultural Festival",
			Status:       "",
			Latitude:     -7.8052,
			Longitude:    110.3647,
			MaxAttendees: 10000,
			TicketPrice:  "Gratis",
			Organizer:    "Keraton Yogyakarta",
			Highlights:   event.JSONArr{"Gamelan Kerajaan", "Pasar Malam Sekaten", "Gunungan Grebeg", "Pertunjukan Wayang"},
		},
		{
			ExternalID:    "prambanan-jazz-2026",
			DestinationID: "prambanan",
			Title:         "Prambanan Jazz Festival 2026",
			Description:  "Prambanan Jazz Festival adalah salah satu festival musik paling ikonik di Indonesia. Digelar di depan latar belakang megah Candi Prambanan yang diterangi cahaya, festival ini menghadirkan musisi jazz terbaik dari Indonesia dan dunia.",
			Location:     "Kompleks Candi Prambanan, Sleman",
			StartDate:    "2026-07-18",
			EndDate:      "2026-07-20",
			ImageURL:     "https://upload.wikimedia.org/wikipedia/commons/4/4f/Prambananjazz-6-2020.png",
			Category:     "Music Festival",
			Status:       "",
			Latitude:     -7.7520,
			Longitude:    110.4914,
			MaxAttendees: 5000,
			TicketPrice:  "IDR 350,000",
			Organizer:    "Rajawali Indonesia",
			Highlights:   event.JSONArr{"Live Jazz Performance", "Sunset Concert", "Heritage Stage", "Culinary Village"},
		},
		{
			ExternalID:    "merapi-adventure-race",
			DestinationID: "merapi",
			Title:         "Merapi Adventure Race 2026",
			Description:  "Tantang dirimu dalam kompetisi petualangan multi-disiplin melintasi medan vulkanik Gunung Merapi. Termasuk trail running, mountain biking, dan navigasi rute. Cocok untuk petualang sejati yang ingin merasakan thrill Merapi.",
			Location:     "Lereng Merapi, Sleman",
			StartDate:    "2026-08-22",
			EndDate:      "2026-08-23",
			ImageURL:     "https://commons.wikimedia.org/wiki/Special:FilePath/Merapi%2C_Yogyakarta.jpg",
			Category:     "Adventure",
			Status:       "",
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
			Status:       "",
			Latitude:     -7.8226,
			Longitude:    110.4107,
			MaxAttendees: 3000,
			TicketPrice:  "IDR 75,000",
			Organizer:    "Dinas Pariwisata DIY",
			Highlights:   event.JSONArr{"Batik Runway Show", "Traditional Weaving Exhibition", "Designer Meet & Greet", "Craft Market"},
		},
		{
			ExternalID:    "parangtritis-kite-festival",
			DestinationID: "parangtritis",
			Title:         "Parangtritis International Kite Festival",
			Description:  "Festival layang-layang internasional yang digelar di pantai ikonik Parangtritis. Ratusan layang-layang raksasa dari berbagai negara menghiasi langit pantai selatan Yogyakarta. Pertunjukan terbaik saat angin laut berhembus kencang.",
			Location:     "Pantai Parangtritis, Bantul",
			StartDate:    "2026-09-19",
			EndDate:      "2026-09-21",
			ImageURL:     "https://images.unsplash.com/photo-1530870110042-98b2cb110834?auto=format&fit=crop&w=1600&q=80",
			Category:     "Cultural Festival",
			Status:       "",
			Latitude:     -8.0257,
			Longitude:    110.3318,
			MaxAttendees: 8000,
			TicketPrice:  "Gratis",
			Organizer:    "Dinas Pariwisata Bantul",
			Highlights:   event.JSONArr{"International Kite Competition", "Night Kite Show", "Beach Sunset View", "Cultural Performance"},
		},
		{
			ExternalID:   "upacara-adat-saparan-bekakak",
			Title:        "Upacara Adat Saparan Bekakak",
			Description:  "Kirab tahunan warga Gamping menyembelih tiruan sepasang pengantin dari ketan berisi kuah gula mawar merah sebagai lambang tolak bala kuno peninggalan Kraton.",
			Location:     "Gunung Gamping, Ambarketawang, Gamping, Sleman",
			StartDate:    "2026-10-01",
			EndDate:      "2026-10-01",
			ImageURL:     "https://upload.wikimedia.org/wikipedia/commons/c/c9/Gunungan_darat_during_Garebeg_Mulud_Yogyakarta_Dec_2017_Pj_IMG_4517sm.jpg",
			Category:     "Cultural Ceremony",
			Status:       "",
			Latitude:     -7.8032,
			Longitude:    110.3212,
			MaxAttendees: 3000,
			TicketPrice:  "Gratis",
			Organizer:    "Keraton Yogyakarta",
			Highlights:   event.JSONArr{"Kirab Adat", "Gunungan Ketan", "Pertunjukan Budaya", "Ritual Tolak Bala"},
		},
		{
			ExternalID:   "upacara-rebo-pungkasan",
			Title:        "Upacara Rebo Pungkasan",
			Description:  "Kirab gunungan lemper raksasa menyambut hari Rabu terakhir bulan Sapar, penanda warisan syukur warga Bantul peninggalan Kyai Wiroyudo pangeran kraton.",
			Location:     "Wonokromo, Pleret, Bantul",
			StartDate:    "2026-11-01",
			EndDate:      "2026-11-01",
			ImageURL:     "https://images.unsplash.com/photo-1514525253161-7a46d19cd819?auto=format&fit=crop&w=1600&q=80",
			Category:     "Cultural Ceremony",
			Status:       "",
			Latitude:     -7.8715,
			Longitude:    110.3951,
			MaxAttendees: 5000,
			TicketPrice:  "Gratis",
			Organizer:    "Pemkab Bantul",
			Highlights:   event.JSONArr{"Gunungan Lemper Raksasa", "Kirab Warga", "Ritual Saparan", "Pasar Rakyat"},
		},
		{
			ExternalID:    "sendratari-roro-jonggrang",
			DestinationID: "prambanan",
			Title:         "Sendratari Roro Jonggrang Prambanan",
			Description:  "Pementasan seni drama tari legendaris kisah berdirinya seribu candi Roro Jonggrang dan Bandung Bondowoso. Digelar megah di panggung terbuka Candi Prambanan dengan tata pencahayaan modern dramatis dan efek multimedia yang memukau serta latar tiga candi utama.",
			Location:     "Kompleks Candi Prambanan, Sleman",
			StartDate:    "2026-07-01",
			EndDate:      "2026-10-31",
			ImageURL:     "https://commons.wikimedia.org/wiki/Special:FilePath/Ramayana_Prambanan_3.jpg",
			Category:     "Cultural Performance",
			Status:       "",
			Latitude:     -7.7520,
			Longitude:    110.4915,
			MaxAttendees: 3000,
			TicketPrice:  "Mulai Rp 50.000",
			Organizer:    "PT Taman Wisata Candi Prambanan",
			Highlights:   event.JSONArr{"Drama Tari Legendaris", "Tata Cahaya Modern", "Latar Candi Prambanan", "Gamelan Langsung"},
		},
		{
			ExternalID:   "ngayogjazz",
			Title:        "Ngayogjazz",
			Description:  "Festival musik jazz tahunan unik berskala internasional yang sengaja digelar di pedesaan (desa wisata) sekitar Yogyakarta. Mengusung konsep merakyat Jazz Mawi Sangu Pari, memadukan keahlian musisi jazz dunia dengan kehangatan ramah tamah warga desa dan kuliner lokal.",
			Location:     "Desa Wisata Pilihan (Berpindah Setiap Tahun di Jogja)",
			StartDate:    "2026-11-14",
			EndDate:      "2026-11-16",
			ImageURL:     "https://images.unsplash.com/photo-1511192336575-5a79af67a629?auto=format&fit=crop&w=1600&q=80",
			Category:     "Music Festival",
			Status:       "",
			Latitude:     -7.6852,
			Longitude:    110.3541,
			MaxAttendees: 5000,
			TicketPrice:  "Gratis (Cukup Bayar Parkir Desa)",
			Organizer:    "Ngayogjazz Committee",
			Highlights:   event.JSONArr{"Jazz di Pedesaan", "Kuliner Lokal Desa", "Panggung Terbuka", "Interaksi Warga Desa"},
		},
		{
			ExternalID:    "seloso-wage-malioboro",
			DestinationID: "malioboro",
			Title:         "Seloso Wage Malioboro",
			Description:  "Hari bebas kendaraan bermotor di sepanjang koridor Jalan Malioboro yang jatuh setiap hari Selasa Wage dalam penanggalan Jawa. Menjadi panggung seni jalanan raksasa di mana ratusan musisi, penari tradisi, badut, dan teater jalanan Jogja tampil estetik menghibur publik.",
			Location:     "Sepanjang Jalan Malioboro, Kota Yogyakarta",
			StartDate:    "2026-07-01",
			EndDate:      "2026-12-31",
			ImageURL:     "https://images.unsplash.com/photo-1543874768-af0b9c4090d5?auto=format&fit=crop&w=1600&q=80",
			Category:     "Street Festival",
			Status:       "",
			Latitude:     -7.7942,
			Longitude:    110.3656,
			MaxAttendees: 10000,
			TicketPrice:  "Gratis",
			Organizer:    "Pemkot Yogyakarta",
			Highlights:   event.JSONArr{"Seni Jalanan", "Musisi Lokal", "Teater Jalanan", "Pasar Rakyat Malioboro"},
		},
		{
			ExternalID:   "kustomfest",
			Title:        "Kustomfest (Indonesia Kustom Kulture Show)",
			Description:  "Festival kustom kulture terbesar di Indonesia yang mempertemukan para modifikator motor, mobil, seniman tato, serta pameran seni retro dunia. Sangat viral di kalangan pecinta otomotif nasional dan global karena menampilkan karya kustom tingkat tinggi nan presisi.",
			Location:     "Jogja Expo Center (JEC), Banguntapan, Bantul",
			StartDate:    "2026-10-03",
			EndDate:      "2026-10-05",
			ImageURL:     "https://images.unsplash.com/photo-1558981806-ec527fa84c39?auto=format&fit=crop&w=1600&q=80",
			Category:     "Custom Culture",
			Status:       "",
			Latitude:     -7.8042,
			Longitude:    110.4005,
			MaxAttendees: 15000,
			TicketPrice:  "Mulai Rp 60.000",
			Organizer:    "Kustomfest Foundation",
			Highlights:   event.JSONArr{"Custom Motorcycle Show", "Tattoo Convention", "Retro Art Exhibition", "Live Music"},
		},
		{
			ExternalID:    "bioskop-sonobudoyo",
			DestinationID: "museum-sonobudoyo",
			Title:         "Bioskop Sonobudoyo",
			Description:  "Pemutaran bioskop alternatif gratis di dalam kompleks Museum Sonobudoyo yang memutarkan berbagai film dokumenter budaya, film pendek karya sineas lokal Jogja, dan sejarah klasik Nusantara. Mengusung bioskop vintage modern mini dengan kualitas audio visual bioskop komersial.",
			Location:     "Gedung Bioskop Museum Sonobudoyo, Kota Yogyakarta",
			StartDate:    "2026-07-01",
			EndDate:      "2026-12-31",
			ImageURL:     "https://upload.wikimedia.org/wikipedia/commons/9/9b/SanabudayaMuseumEntrySign85.jpg",
			Category:     "Film & Cinema",
			Status:       "",
			Latitude:     -7.8025,
			Longitude:    110.3635,
			MaxAttendees: 100,
			TicketPrice:  "Gratis (Wajib Reservasi Online)",
			Organizer:    "Museum Sonobudoyo",
			Highlights:   event.JSONArr{"Film Dokumenter Budaya", "Film Pendek Sineas Lokal", "Bioskop Vintage", "Diskusi Seni"},
		},
		{
			ExternalID:    "sendratari-ramayana-prambanan",
			DestinationID: "prambanan",
			Title:         "Sendratari Ramayana Prambanan",
			Description:  "Seni pertunjukan tari teater paling spektakuler dan legendaris di Indonesia yang mementaskan kisah kepahlawanan Rama menyelamatkan Shinta. Ditampilkan tanpa dialog, mengandalkan ekspresi gerak tari Jawa indah, gamelan rancak, dan latar belakang megah tiga candi utama Prambanan yang bersinar keemasan.",
			Location:     "Panggung Terbuka Candi Prambanan, Sleman",
			StartDate:    "2026-05-01",
			EndDate:      "2026-10-31",
			ImageURL:     "https://commons.wikimedia.org/wiki/Special:FilePath/Ramayana_Prambanan_3.jpg",
			Category:     "Cultural Performance",
			Status:       "",
			Latitude:     -7.7512,
			Longitude:    110.4901,
			MaxAttendees: 3000,
			TicketPrice:  "Mulai Rp 150.000",
			Organizer:    "PT Taman Wisata Candi Prambanan",
			Highlights:   event.JSONArr{"Tari Teater Tanpa Dialog", "Latar Candi Prambanan", "Gamelan Tradisional", "Kisah Rama-Shinta"},
		},
		{
			ExternalID:   "yogyakarta-gamelan-festival",
			Title:        "Yogyakarta Gamelan Festival (YGF)",
			Description:  "Gathering seni gamelan internasional legendaris yang mempertemukan para pemain, pencipta, dan pecinta gamelan dari seluruh penjuru dunia. Menyajikan pementasan gamelan kontemporer radikal hingga klasik purba yang sangat viral mendunia.",
			Location:     "Gedung TBY atau Pendopo Heritage di Yogyakarta",
			StartDate:    "2026-09-05",
			EndDate:      "2026-09-08",
			ImageURL:     "https://commons.wikimedia.org/wiki/Special:FilePath/Sarons_of_Gamelan_Sekati,_Yogyakarta.jpg",
			Category:     "Music Festival",
			Status:       "",
			Latitude:     -7.8005,
			Longitude:    110.3685,
			MaxAttendees: 2000,
			TicketPrice:  "Gratis (Sebagian Berbayar)",
			Organizer:    "YGF Foundation",
			Highlights:   event.JSONArr{"Gamelan Internasional", "Workshop Gamelan", "Pertunjukan Kolosal", "Pertemuan Seniman Dunia"},
		},
		{
			ExternalID:    "jogja-international-street-performance",
			DestinationID: "malioboro",
			Title:         "Jogja International Street Performance (JISP)",
			Description:  "Festival tari jalanan berskala internasional yang membawa panggung seni keluar dari gedung teater menuju ruang publik terbuka di Yogyakarta. Ratusan seniman tari jalanan mancanegara menampilkan ekspresi gerak tubuh kreatif interaktif di hadapan publik Malioboro.",
			Location:     "Jalan Malioboro dan Titik Nol Kilometer, Kota Yogyakarta",
			StartDate:    "2026-10-20",
			EndDate:      "2026-10-22",
			ImageURL:     "https://images.unsplash.com/photo-1508700115892-45ecd05ae2ad?auto=format&fit=crop&w=1600&q=80",
			Category:     "Street Performance",
			Status:       "",
			Latitude:     -7.8012,
			Longitude:    110.3648,
			MaxAttendees: 8000,
			TicketPrice:  "Gratis",
			Organizer:    "Dinas Pariwisata DIY",
			Highlights:   event.JSONArr{"Tari Jalanan Internasional", "Seniman Mancanegara", "Interaktif Publik", "Panggung Terbuka Malioboro"},
		},
		{
			ExternalID:    "sandyakala-jonggrang",
			DestinationID: "ratuboko",
			Title:         "Sandyakala Jonggrang",
			Description:  "Pertunjukan seni budaya magis yang menghidupkan legenda Roro Jonggrang di tengah megahnya Keraton Ratu Boko. Dimulai saat senja pukul 17.00 WIB, perpaduan tarian tradisional, gamelan, dan latar sunset yang memukau menciptakan pengalaman budaya yang tak terlupakan. Diselenggarakan oleh PT Taman Wisata Candi Borobudur, Prambanan, dan Ratu Boko.",
			Location:     "Keraton Ratu Boko, Sleman",
			StartDate:    "2026-06-21",
			EndDate:      "2026-07-03",
			ImageURL:     "https://images.unsplash.com/photo-1508700115892-45ecd05ae2ad?auto=format&fit=crop&w=1600&q=80",
			Category:     "Cultural Performance",
			Status:       "",
			Latitude:     -7.7700,
			Longitude:    110.4880,
			MaxAttendees: 2000,
			TicketPrice:  "Gratis (dengan tiket masuk Ratu Boko Rp40.000)",
			Organizer:    "PT Taman Wisata Candi Borobudur, Prambanan, dan Ratu Boko",
			Highlights:   event.JSONArr{"Tarian Roro Jonggrang", "Sunset di Ratu Boko", "Gamelan Tradisional", "Pertunjukan Senja"},
		},
		{
			ExternalID:    "centili-obelix-hills",
			DestinationID: "obelix-hills",
			Title:         "Centili Dance Group di Obelix Hills",
			Description:  "Centili Dance Group adalah grup tari komedi yang sangat populer di Yogyakarta. Menyajikan pertunjukan tari unik yang memadukan unsur komedi, kostum nyentrik, dan gerakan akrobatik. Para penari pria sering memerankan karakter wanita dengan gaya energik dan menghibur. Tampil setiap hari Sabtu di Obelix Hills Prambanan mulai pukul 16.00 WIB.",
			Location:     "Obelix Hills, Klumprit, Prambanan, Sleman",
			StartDate:    "2026-07-13",
			EndDate:      "2026-07-13",
			ImageURL:     "https://images.unsplash.com/photo-1545156521-77bd85671d30?auto=format&fit=crop&w=1600&q=80",
			Category:     "Dance & Comedy",
			Status:       "",
			Latitude:     -7.7612,
			Longitude:    110.5063,
			MaxAttendees: 500,
			TicketPrice:  "Gratis (dengan tiket masuk Obelix Hills Rp20.000-Rp25.000)",
			Organizer:    "Obelix Hills Management",
			Highlights:   event.JSONArr{"Tari Komedi", "Kostum Nyentrik", "Akrobatik", "Hiburan Keluarga"},
		},
		{
			ExternalID:    "centili-obelix-sea-view",
			DestinationID: "obelix-sea-view",
			Title:         "Centili Dance Group di Obelix Sea View",
			Description:  "Centili Dance Group adalah grup tari komedi yang sangat populer di Yogyakarta. Menyajikan pertunjukan tari unik yang memadukan unsur komedi, kostum nyentrik, dan gerakan akrobatik. Para penari pria sering memerankan karakter wanita dengan gaya energik dan menghibur. Tampil rutin di Obelix Sea View Gunungkidul pada pukul 17.00 WIB. Cek jadwal terbaru di Instagram @obelixseaview.",
			Location:     "Obelix Sea View, Giricahyo, Purwosari, Gunungkidul",
			StartDate:    "2026-07-01",
			EndDate:      "2026-07-31",
			ImageURL:     "https://images.unsplash.com/photo-1545156521-77bd85671d30?auto=format&fit=crop&w=1600&q=80",
			Category:     "Dance & Comedy",
			Status:       "",
			Latitude:     -8.0285,
			Longitude:    110.3325,
			MaxAttendees: 800,
			TicketPrice:  "Gratis (dengan tiket masuk Obelix Sea View Rp30.000-Rp50.000)",
			Organizer:    "Obelix Sea View Management",
			Highlights:   event.JSONArr{"Tari Komedi", "Kostum Nyentrik", "Akrobatik", "Hiburan Keluarga"},
		},
	}

	inserted := 0
	updated := 0
	for _, e := range events {
		// Auto-detect status from dates if not set
		if e.Status == "" {
			e.Status = resolveEventStatus(e.StartDate, e.EndDate)
		}

		existing := event.Event{}
		if err := db.Where("external_id = ?", e.ExternalID).First(&existing).Error; err == nil {
			e.ID = existing.ID
			e.CreatedAt = existing.CreatedAt
			if err := db.Save(&e).Error; err != nil {
				log.Printf("Failed to update event %s: %v", e.ExternalID, err)
				continue
			}
			updated++
			continue
		}
		if err := db.Create(&e).Error; err != nil {
			log.Printf("Failed to seed event %s: %v", e.ExternalID, err)
			continue
		}
		inserted++
	}
	fmt.Printf("Events seeding done: %d inserted, %d updated\n", inserted, updated)
}

func resolveEventStatus(startDate, endDate string) string {
	now := time.Now()
	layout := "2006-01-02"

	if startDate != "" {
		if start, err := time.Parse(layout, startDate); err == nil && start.After(now) {
			return "upcoming"
		}
	}
	if endDate != "" {
		if end, err := time.Parse(layout, endDate); err == nil && end.Before(now) {
			return "completed"
		}
	}
	if startDate != "" {
		if _, err := time.Parse(layout, startDate); err == nil {
			return "active"
		}
	}
	return "upcoming"
}
