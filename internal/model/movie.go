package model

import "diplomM/internal/model/kinopoisk"

// MovieSearchResponse - ответ поиска фильмов
type MovieSearchResponse struct {
	Docs  []kinopoisk.KinopoiskFilmSearchItem `json:"docs"`
	Total int                                 `json:"total"`
	Limit int                                 `json:"limit"`
	Page  int                                 `json:"page"`
	Pages int                                 `json:"pages"`
}

// MoviePreview - превью фильма (для похожих)
type MoviePreview struct {
	FilmID           int64    `json:"filmId"`
	NameRU           *string  `json:"nameRu"`
	NameEN           *string  `json:"nameEn"`
	PosterURL        string   `json:"posterUrl"`
	PosterURLPreview string   `json:"posterUrlPreview"`
	RatingKinopoisk  *float64 `json:"ratingKinopoisk"`
	RatingIMDB       *float64 `json:"ratingImdb"`
	Year             *int     `json:"year"`
	Type             string   `json:"type"`
}
