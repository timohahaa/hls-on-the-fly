package storage

type Asset struct {
	Quality    string
	Resolution string
	FPS        float64
	// считаем, что кодек мы знаем, но при необходимости кодек можно достать налету из файла
	Codec    string
	Duration float64
	FilePath string
}
