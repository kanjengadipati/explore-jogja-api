package seeds

import (
	"encoding/json"
	"fmt"
	"log"

	"pleco-api/internal/modules/destination"

	"gorm.io/gorm"
)

func SeedDestinations(db *gorm.DB) {
	mustHaveDB(db)

	var count int64
	db.Model(&destination.Destination{}).Count(&count)
	if count > 0 {
		fmt.Println("Destinations already seeded, skipping")
		return
	}

	dests := getDestinationSeedData()
	if err := db.CreateInBatches(dests, 10).Error; err != nil {
		log.Printf("Failed to seed destinations: %v", err)
		return
	}
	fmt.Printf("Destinations seeding done (%d records)\n", len(dests))
}

func ja(v interface{}) destination.JSONArr {
	b, _ := json.Marshal(v)
	var arr destination.JSONArr
	json.Unmarshal(b, &arr)
	return arr
}

func jm(v interface{}) destination.JSONMap {
	b, _ := json.Marshal(v)
	var m destination.JSONMap
	json.Unmarshal(b, &m)
	return m
}

func getDestinationSeedData() []destination.Destination {
	return []destination.Destination{
		{
			ExternalID:  "prambanan",
			Name:        "Prambanan Temple",
			Tagline:     "The Pinnacle of Royal Hindu Architecture",
			Category:    "heritage",
			Location:    "Sleman, Yogyakarta",
			SubRegion:   "Sleman",
			Images:      ja([]string{"https://images.unsplash.com/photo-1584810359583-96fc3448beaa?q=80&w=1200", "https://images.unsplash.com/photo-1626847037657-fd3622613ce3?q=80&w=1200", "https://images.unsplash.com/photo-1601999109332-542b18dbec57?q=80&w=1200"}),
			Rating:      4.8,
			ReviewCount: 3840,
			Description: "Built in the 9th century, Prambanan is the largest Hindu temple compound in Indonesia, dedicated to the Trimurti: the Creator (Brahma), the Preserver (Vishnu), and the Destroyer (Shiva). The towering spires reach 47 meters high, dominating the surrounding plains.",
			Story:       "Legend tells of Roro Jonggrang, a beautiful princess who demanded Prince Bandung Bondowoso build 1,000 temples in a single night to win her hand. With the help of spirits, Bandung nearly succeeded, but Roro Jonggrang tricked the roosters into crowing early. Realizing the deception, Bandung cursed her to become the final, 1,000th stone statue—which sits in the Shiva chamber to this day.",
			TicketPrice: "IDR 375,000 (Foreigners) / IDR 50,000 (Domestic)",
			OpeningHours: "06:30 AM - 05:00 PM Daily",
			Facilities:  ja([]string{"Visitor Information Center", "Audio Guides", "Wheelchair Access", "Spacious Parking", "Art Souvenir Arcades", "Traditional Restaurants", "Clean Restrooms"}),
			TravelTips:  ja([]string{"Visit in the late afternoon (around 03:30 PM) to capture the beautiful golden hour light shining through the spires.", "Wear modest clothing out of respect for the sacred site. Sarongs are provided at the entrance.", "Check out the Ramayana Ballet performance scheduled on open-air stages during dry season nights."}),
			BestTime:    "May to October (Dry Season, ideal for clear sunset backdrops)",
			Weather:     jm(map[string]string{"temp": "28°C", "condition": "Sunny", "status": "Perfect weather for exploring Prambanan today."}),
			Latitude:    -7.7520,
			Longitude:   110.4914,
			Reviews:     ja([]map[string]interface{}{{"id": "r1", "userName": "Sophia Laurent", "userAvatar": "https://images.unsplash.com/photo-1494790108377-be9c29b29330?q=80&w=150", "rating": 5, "date": "2026-06-15", "comment": "An absolutely magical experience."}, {"id": "r2", "userName": "Yuki Tanaka", "userAvatar": "https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?q=80&w=150", "rating": 4.8, "date": "2026-07-02", "comment": "Watching the sunset over these structures is a lifetime memory."}}),
			Partners:    ja([]map[string]interface{}{{"id": "p-p1", "name": "The Phoenix Hotel Yogyakarta", "category": "hotel", "image": "https://images.unsplash.com/photo-1566073771259-6a8506099945?q=80&w=600", "rating": 4.9, "price": "IDR 1,500,000 / night", "distance": "14 km from Prambanan", "description": "A luxurious landmark colonial boutique heritage hotel.", "address": "Jl. Jend. Sudirman No.9, Yogyakarta"}}),
			FAQs:        ja([]map[string]interface{}{{"q": "Is there a dress code?", "a": "Yes. Modest wear covering knees and shoulders is recommended."}, {"q": "When does the Ramayana Ballet perform?", "a": "It performs on Tuesday, Thursday, and Saturday nights."}}),
		},
		{
			ExternalID:  "malioboro",
			Name:        "Malioboro Street",
			Tagline:     "The Soul and Lifeline of Yogyakarta",
			Category:    "culinary",
			Location:    "Yogyakarta City",
			SubRegion:   "Yogyakarta",
			Images:      ja([]string{"https://images.unsplash.com/photo-1621360841013-c7683c659ec6?q=80&w=1200", "https://images.unsplash.com/photo-1581456495146-65a71b2c8e52?q=80&w=1200", "https://images.unsplash.com/photo-1596402184320-417e7178b2cd?q=80&w=1200"}),
			Rating:      4.6,
			ReviewCount: 9280,
			Description: "The primary shopping street of Yogyakarta, pulsing with energy day and night. Framed by colonial-era storefronts, traditional horse carriages (Andong), motorized three-wheelers (Becak), and street musicians.",
			Story:       "Malioboro represents the imaginary axis linking Mount Merapi in the north, the Sultan's Palace (Kraton) in the center, and the mystical South Sea (Parangtritis) in the south.",
			TicketPrice: "Free Entry",
			OpeningHours: "24 Hours Daily (Best at evening 06:00 PM - 11:00 PM)",
			Facilities:  ja([]string{"Pedestrian Benches", "Batik Stores", "Food Courtyards", "Street Musicians", "Trishaw Stands", "Historic Buildings", "Tourist Police Centers"}),
			TravelTips:  ja([]string{"Take a slow evening walk starting from Tugu Station down to the central post office.", "Practice friendly, smiling negotiation when purchasing handmade batik.", "Try dining at an authentic open-floor bamboo mat stall (Lesehan) serving traditional Gudeg."}),
			BestTime:    "Every evening after 06:00 PM",
			Weather:     jm(map[string]string{"temp": "27°C", "condition": "Clear Evening", "status": "Perfect cool breeze for an evening walk in Malioboro."}),
			Latitude:    -7.7928,
			Longitude:   110.3658,
			Reviews:     ja([]map[string]interface{}{{"id": "r1", "userName": "Sophia Laurent", "userAvatar": "https://images.unsplash.com/photo-1494790108377-be9c29b29330?q=80&w=150", "rating": 5, "date": "2026-06-15", "comment": "An absolutely magical experience."}}),
			Partners:    ja([]map[string]interface{}{{"id": "m-p1", "name": "Grand Inna Malioboro Heritage", "category": "hotel", "image": "https://images.unsplash.com/photo-1551882547-ff40c63fe5fa?q=80&w=600", "rating": 4.7, "price": "IDR 1,100,000 / night", "distance": "0 km", "description": "An iconic historic heritage hotel established in 1908.", "address": "Jl. Malioboro No.60, Yogyakarta"}}),
			FAQs:        ja([]map[string]interface{}{{"q": "Is Malioboro wheelchair accessible?", "a": "Yes. The sidewalks are exceptionally wide."}, {"q": "What is the best way to get there?", "a": "Arriving by train at Tugu Station places you directly at the northern tip."}}),
		},
		{
			ExternalID:  "parangtritis",
			Name:        "Parangtritis Beach",
			Tagline:     "Mystical Golden Sands of the Southern Realm",
			Category:    "beach",
			Location:    "Bantul, Yogyakarta",
			SubRegion:   "Bantul",
			Images:      ja([]string{"https://images.unsplash.com/photo-1542314831-068cd1dbfeeb?q=80&w=1200", "https://images.unsplash.com/photo-1507525428034-b723cf961d3e?q=80&w=1200", "https://images.unsplash.com/photo-1519046904884-53103b34b206?q=80&w=1200"}),
			Rating:      4.7,
			ReviewCount: 4230,
			Description: "Framed by dramatic black volcanic sands, towering karst cliffs, and roaring waves from the Indian Ocean, Parangtritis is Yogyakarta's most legendary beach.",
			Story:       "Parangtritis is deeply woven with Javanese cosmology. It is believed to be the sacred gateway to the undersea palace of Kanjeng Ratu Kidul, the mystical Queen of the Southern Seas.",
			TicketPrice: "IDR 15,000 / person",
			OpeningHours: "24 Hours Daily",
			Facilities:  ja([]string{"Traditional Chariot Rides", "ATV Rentals", "Fresh Coconut Stalls", "Cliffside Gazebos", "Volcanic Sand Dunes", "Lifeguard Posts", "Local Seafood Dining"}),
			TravelTips:  ja([]string{"Rent a horse-drawn carriage (Andong) to gallop along the tideline during sunset.", "Head to the nearby Gumuk Pasir for sandboarding.", "Do not swim in the ocean. The undercurrents are extremely powerful."}),
			BestTime:    "June to August, when sunsets are exceptionally crisp.",
			Weather:     jm(map[string]string{"temp": "29°C", "condition": "Ocean Breeze", "status": "Beautiful clear skies over the South Sea today."}),
			Latitude:    -8.0253,
			Longitude:   110.3298,
			Reviews:     ja([]map[string]interface{}{{"id": "r1", "userName": "Sophia Laurent", "userAvatar": "https://images.unsplash.com/photo-1494790108377-be9c29b29330?q=80&w=150", "rating": 5, "date": "2026-06-15", "comment": "An absolutely magical experience."}}),
			Partners:    ja([]map[string]interface{}{{"id": "pt-p1", "name": "Queen of the South Resort", "category": "hotel", "image": "https://images.unsplash.com/photo-1571896349842-33c89424de2d?q=80&w=600", "rating": 4.8, "price": "IDR 1,400,000 / night", "distance": "1.5 km from beach", "description": "Perched high on the cliffs with an infinity pool.", "address": "Parangrejo, Purwosari, Bantul"}}),
			FAQs:        ja([]map[string]interface{}{{"q": "Can I swim in the sea?", "a": "Strictly prohibited."}, {"q": "Is there sandboarding?", "a": "Yes! At Gumuk Pasir Parangkusumo."}}),
		},
		{
			ExternalID:  "merapi",
			Name:        "Mount Merapi Lava Tour",
			Tagline:     "An Unforgettable Offroad Journey on an Active Volcano",
			Category:    "adventure",
			Location:    "Sleman, Yogyakarta",
			SubRegion:   "Sleman",
			Images:      ja([]string{"https://images.unsplash.com/photo-1544735716-392fe2489ffa?q=80&w=1200", "https://images.unsplash.com/photo-1604999333679-b86d54738315?q=80&w=1200", "https://images.unsplash.com/photo-1511497584788-876760111969?q=80&w=1200"}),
			Rating:      4.8,
			ReviewCount: 5120,
			Description: "Ride inside open-cabin vintage 4x4 Willys Jeeps along the trails left by Mount Merapi's historical eruptions.",
			Story:       "Mount Merapi is one of the world's most active volcanoes. The local Javanese hold deep reverence for the mountain spirits.",
			TicketPrice: "IDR 350,000 - IDR 650,000 per Jeep",
			OpeningHours: "04:30 AM - 06:00 PM Daily",
			Facilities:  ja([]string{"Willys Jeep Guild", "Licensed Offroad Drivers", "Protective Helmets & Masks", "Merapi Relic Museum", "Underground Kaliadem Bunker", "Scenic Mountain View Cafes"}),
			TravelTips:  ja([]string{"The Sunrise Jeep Tour is the best experience.", "Bring a light windbreaker jacket.", "Wear the provided face mask for dusty trails."}),
			BestTime:    "Dry months from May to September.",
			Weather:     jm(map[string]string{"temp": "22°C", "condition": "Mist & Sunrise", "status": "Stunning cloudless view of the volcano summit."}),
			Latitude:    -7.5960,
			Longitude:   110.4463,
			Reviews:     ja([]map[string]interface{}{{"id": "r1", "userName": "Sophia Laurent", "userAvatar": "https://images.unsplash.com/photo-1494790108377-be9c29b29330?q=80&w=150", "rating": 5, "date": "2026-06-15", "comment": "An absolutely magical experience."}}),
			Partners:    ja([]map[string]interface{}{{"id": "me-p1", "name": "Kopi Klotok Pakem", "category": "restaurant", "image": "https://images.unsplash.com/photo-1501339847302-ac426a4a7cbb?q=80&w=600", "rating": 4.9, "price": "IDR 20,000 - 45,000 / person", "distance": "6.5 km from Tour Base", "description": "Rustic traditional Javanese village cafe.", "address": "Jl. Kaliurang Km 16, Pakem, Sleman"}}),
			FAQs:        ja([]map[string]interface{}{{"q": "Is it safe for kids?", "a": "Yes. Families can choose the Short Route."}, {"q": "What is inside the Bunker?", "a": "A steel-reinforced shelter built to withstand volcanic heat."}}),
		},
		{
			ExternalID:  "tamansari",
			Name:        "Taman Sari Water Castle",
			Tagline:     "The Secret Royal Bathing Pools of the Sultanate",
			Category:    "heritage",
			Location:    "Yogyakarta City",
			SubRegion:   "Yogyakarta",
			Images:      ja([]string{"https://images.unsplash.com/photo-1581456495146-65a71b2c8e52?q=80&w=1200", "https://images.unsplash.com/photo-1596402184320-417e7178b2cd?q=80&w=1200", "https://images.unsplash.com/photo-1584810359583-96fc3448beaa?q=80&w=1200"}),
			Rating:      4.6,
			ReviewCount: 3120,
			Description: "Built in the mid-18th century as a private pleasure park for the first Sultan, Taman Sari is a stunning architectural mixture of Javanese and Portuguese styles.",
			Story:       "Sultan Hamengkubuwono I built Taman Sari as a resting palace, defense castle, and mystical meditation sanctuary.",
			TicketPrice: "IDR 15,000 / person",
			OpeningHours: "09:00 AM - 03:30 PM Daily",
			Facilities:  ja([]string{"English Speaking Royal Guides", "Turquoise Bathing Pools", "Underground Mosque", "Passage Tunnels", "Heritage Batik Villages", "Craft Souvenir Galleries"}),
			TravelTips:  ja([]string{"Hire a local guide to navigate the labyrinth of pastel narrow streets.", "Excellent photo opportunities at the central five-staircase mosque.", "Visit the nearby Kampung Cyber."}),
			BestTime:    "Morning between 09:30 AM - 11:30 AM",
			Weather:     jm(map[string]string{"temp": "28°C", "condition": "Partly Cloudy", "status": "Warm sunshine, perfect for the outdoor pools today."}),
			Latitude:    -7.8101,
			Longitude:   110.3592,
			Reviews:     ja([]map[string]interface{}{{"id": "r1", "userName": "Sophia Laurent", "userAvatar": "https://images.unsplash.com/photo-1494790108377-be9c29b29330?q=80&w=150", "rating": 5, "date": "2026-06-15", "comment": "An absolutely magical experience."}}),
			Partners:    ja([]map[string]interface{}{{"id": "ts-p1", "name": "The Phoenix Hotel Yogyakarta", "category": "hotel", "image": "https://images.unsplash.com/photo-1566073771259-6a8506099945?q=80&w=600", "rating": 4.9, "price": "IDR 1,500,000 / night", "distance": "4.2 km from Taman Sari", "description": "Colonial elegance blended with royal Javanese architecture.", "address": "Jl. Jend. Sudirman No.9, Yogyakarta"}}),
			FAQs:        ja([]map[string]interface{}{{"q": "Is Sumur Gumuling open?", "a": "Yes."}, {"q": "Can we photograph the pools?", "a": "Yes! Personal photography is permitted and free."}}),
		},
		{
			ExternalID:  "goajomblang",
			Name:        "Goa Jomblang Cave",
			Tagline:     "The Celestial Beam of Heavenly Light",
			Category:    "hidden-gem",
			Location:    "Gunungkidul, Yogyakarta",
			SubRegion:   "Gunungkidul",
			Images:      ja([]string{"https://images.unsplash.com/photo-1604999333679-b86d54738315?q=80&w=1200", "https://images.unsplash.com/photo-1511497584788-876760111969?q=80&w=1200", "https://images.unsplash.com/photo-1542314831-068cd1dbfeeb?q=80&w=1200"}),
			Rating:      4.9,
			ReviewCount: 1840,
			Description: "Goa Jomblang is a vertical collapse-sinkhole cave. You will be rappelled down 60 meters into a mystical underground ancient forest.",
			Story:       "Formed hundreds of thousands of years ago when the ground collapsed into a vertical limestone sinkhole.",
			TicketPrice: "IDR 500,000 / person",
			OpeningHours: "07:30 AM - 12:30 PM",
			Facilities:  ja([]string{"Certified SRT Guides", "Rappelling Harnesses", "Rubber Boots", "Outdoor Shower Rooms", "Traditional Lunch Box", "Resting Gazebos"}),
			TravelTips:  ja([]string{"Advance bookings are mandatory (80 people/day limit).", "Bring dry spare clothes and a towel.", "Pack a bright headlamp."}),
			BestTime:    "Sunny days between 11:15 AM and 12:30 PM",
			Weather:     jm(map[string]string{"temp": "27°C", "condition": "Sunny Outside", "status": "Perfect clear skies today!"}),
			Latitude:    -8.0287,
			Longitude:   110.6384,
			Reviews:     ja([]map[string]interface{}{{"id": "r1", "userName": "Sophia Laurent", "userAvatar": "https://images.unsplash.com/photo-1494790108377-be9c29b29330?q=80&w=150", "rating": 5, "date": "2026-06-15", "comment": "An absolutely magical experience."}}),
			Partners:    ja([]map[string]interface{}{{"id": "gj-p1", "name": "Gunungkidul Cave Guides Guild", "category": "guide", "image": "https://images.unsplash.com/photo-1544717297-fa95b6ee9643?q=80&w=600", "rating": 4.9, "price": "Included in entry ticket", "distance": "On-Site", "description": "Highly trained single-rope safety experts.", "address": "Pacarejo, Semanu, Gunungkidul"}}),
			FAQs:        ja([]map[string]interface{}{{"q": "Do I need caving experience?", "a": "No."}, {"q": "Are there age restrictions?", "a": "Yes. Recommended for ages 10-60."}}),
		},
		{
			ExternalID:  "kalibiru",
			Name:        "Kalibiru National Park",
			Tagline:     "Highland Serenity and Panoramic Reservoir Views",
			Category:    "nature",
			Location:    "Kulon Progo, Yogyakarta",
			SubRegion:   "Kulon Progo",
			Images:      ja([]string{"https://images.unsplash.com/photo-1501785888041-af3ef285b470?q=80&w=1200"}),
			Rating:      4.7,
			ReviewCount: 2180,
			Description: "Kalibiru National Park is located in the Menoreh Hills, offering a breathtaking panoramic view of the Sermo Reservoir and lush green pine forests.",
			Story:       "Kalibiru was originally a dry and degraded forest area. Through collaborative reforestation efforts, it blossomed into a pioneering eco-tourism model.",
			TicketPrice: "IDR 20,000 / person",
			OpeningHours: "08:00 AM - 05:00 PM Daily",
			Facilities:  ja([]string{"Treetop Viewing Decks", "Flying Fox Ziplines", "Trekking Trails", "Local Viewpoint Cafes", "Spacious Parking", "Clean Restrooms"}),
			TravelTips:  ja([]string{"Arrive early for the coolest mountain air.", "Wear comfortable hiking shoes.", "Hire a local photographer at the platforms."}),
			BestTime:    "Dry season months of June to September",
			Weather:     jm(map[string]string{"temp": "24°C", "condition": "Clear & Cool", "status": "Beautiful clear skies with a cool breeze."}),
			Latitude:    -7.8052,
			Longitude:   110.1293,
			Reviews:     ja([]map[string]interface{}{{"id": "r1", "userName": "Sophia Laurent", "userAvatar": "https://images.unsplash.com/photo-1494790108377-be9c29b29330?q=80&w=150", "rating": 5, "date": "2026-06-15", "comment": "An absolutely magical experience."}}),
			Partners:    ja([]map[string]interface{}{{"id": "kb-p1", "name": "Sermo Lakeside Camping", "category": "rental", "image": "https://images.unsplash.com/photo-1504280390367-361c6d9f38f4?q=80&w=600", "rating": 4.7, "price": "IDR 150,000 / night", "distance": "4.5 km from Kalibiru", "description": "Camp under the stars on the edge of the reservoir.", "address": "Sermo Reservoir Area, Kulon Progo"}}),
			FAQs:        ja([]map[string]interface{}{{"q": "Is it safe to climb the platforms?", "a": "Yes. Every platform is equipped with safety harnesses."}}),
		},
	}
}
