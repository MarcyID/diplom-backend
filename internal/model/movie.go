package model

type Movie struct {
	ID               int64          `json:"id"`
	Name             string         `json:"name"`
	AlternativeName  string         `json:"alternativeName"`
	EnName           string         `json:"enName"`
	Type             string         `json:"type"` // movie, tv-series
	Year             int            `json:"year"`
	Description      string         `json:"description"`
	ShortDescription string         `json:"shortDescription"`
	Rating           Rating         `json:"rating"`
	Votes            Votes          `json:"votes"`
	MovieLength      int            `json:"movieLength"`
	AgeRating        int            `json:"ageRating"`
	Poster           Image          `json:"poster"`
	Backdrop         Image          `json:"backdrop"`
	Genres           []Genre        `json:"genres"`
	Countries        []Country      `json:"countries"`
	Persons          []Person       `json:"persons"`
	Watchability     Watchability   `json:"watchability"`
	SimilarMovies    []MoviePreview `json:"similarMovies"`
}

type Rating struct {
	KP   float64 `json:"kp"`
	IMDB float64 `json:"imdb"`
	TMDB float64 `json:"tmdb"`
}

type Votes struct {
	KP   int `json:"kp"`
	IMDB int `json:"imdb"`
}

type Image struct {
	URL        string `json:"url"`
	PreviewURL string `json:"previewUrl"`
}

type Genre struct {
	Name string `json:"name"`
}

type Country struct {
	Name string `json:"name"`
}

type Person struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Profession string `json:"profession"`
	Photo      string `json:"photo"`
}

type Watchability struct {
	Items []WatchItem `json:"items"`
}

type WatchItem struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type MovieSearchResponse struct {
	Docs  []Movie `json:"docs"`
	Total int     `json:"total"`
	Limit int     `json:"limit"`
	Page  int     `json:"page"`
	Pages int     `json:"pages"`
}

type MoviePreview struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Year   int    `json:"year"`
	Rating Rating `json:"rating"`
	Poster Image  `json:"poster"`
}
