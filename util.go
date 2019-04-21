package main

type ServiceMeta struct {
	CurrentPageNumber  int    `json:"current_page_number"`
	ErrorText          string `json:"error_text"`
	NextPageNumber     int    `json:"next_page_number"`
	NumberOfPages      int    `json:"number_of_pages"`
	PreviousPageNumber int    `json:"previous_page_number"`
	ResultsPerPage     int    `json:"results_per_page"`
}