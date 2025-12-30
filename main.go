package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	godotenv "github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

type WeatherData struct {
	Location    string
	Temperature int
	Condition   string
	Humidity    int
	WindSpeed   float32
	FeelsLike   int
}

type OpenWeatherResponse struct {
	Name string `json:"name"`
	Main struct {
		Temp      float32 `json:"temp"`
		FeelsLike float32 `json:"feels_like"`
		Humidity  int     `json:"humidity"`
	} `json:"main"`
	Weather []struct {
		Main        string `json:"main"`
		Description string `json:"description"`
	} `json:"weather"`
	Wind struct {
		Speed float32 `json:"speed"`
	} `json:"wind"`
}

// FETCH WEATHER DATA FUNCTION
func fetchWeatherData(cityName string) (WeatherData, error) {
	var weather WeatherData

	apiKey := os.Getenv("API_KEY")
	apiURL := os.Getenv("API_URL")

	url := fmt.Sprintf("%s?q=%s&appid=%s&units=metric", apiURL, cityName, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return weather, fmt.Errorf("failed to fetch weather: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return weather, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return weather, fmt.Errorf("failed to read response: %v", err)
	}

	var apiResp map[string]interface{}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return weather, fmt.Errorf("failed to parse JSON: %v", err)
	}

	weather = WeatherData{
		Location:    apiResp["name"].(string),
		Temperature: int(apiResp["main"].(map[string]interface{})["temp"].(float64)),
		FeelsLike:   int(apiResp["main"].(map[string]interface{})["feels_like"].(float64)),
		Humidity:    int(apiResp["main"].(map[string]interface{})["humidity"].(float64)),
		WindSpeed:   float32(apiResp["wind"].(map[string]interface{})["speed"].(float64)),
	}

	return weather, nil
}

func main() {

	const (
		WIDTH           int32  = 800
		HEIGHT          int32  = 450
		FPS             int32  = 60
		MAX_INPUT_CHARS int    = 18
		FONT_PATH       string = "resource/static/JetBrainsMono-Regular.ttf"
	)

	var (
		name            = make([]rune, MAX_INPUT_CHARS+1)
		letterCount     int
		framesCounter   int
		mouseOnText     bool
		textBox         rl.Rectangle
		inputText       string
		statusMessage   string
		statusColor     rl.Color
		statusClearTime time.Time
		weather         WeatherData
		lastFetchTime   time.Time
		fetchCooldown   = 2 * time.Second
	)

	rl.InitWindow(WIDTH, HEIGHT, "Go Weather")
	defer rl.CloseWindow()

	rl.SetTargetFPS(FPS)

	font := rl.LoadFontEx(FONT_PATH, 48, nil)
	defer rl.UnloadFont(font)

	rl.SetTextureFilter(font.Texture, rl.FilterBilinear)

	//  INIT TEXTBOX RECTANGLE
	textBox = rl.NewRectangle(225, 80, 350, 50)

	for !rl.WindowShouldClose() {

		// UPDATE
		if rl.CheckCollisionPointRec(rl.GetMousePosition(), textBox) {
			mouseOnText = true
		} else {
			mouseOnText = false
		}

		if mouseOnText {

			rl.SetMouseCursor(rl.MouseCursorIBeam)

			key := rl.GetCharPressed()

			for key > 0 {

				if key >= 32 && key <= 125 && letterCount < MAX_INPUT_CHARS {

					name[letterCount] = rune(key)
					letterCount++

					if letterCount < len(name) {
						name[letterCount] = 0
					}
				}

				key = rl.GetCharPressed()
			}

			if rl.IsKeyPressed(rl.KeyBackspace) {
				letterCount--
				if letterCount < 0 {
					letterCount = 0
				}
				name[letterCount] = 0
			}

		} else {
			rl.SetMouseCursor(rl.MouseCursorDefault)
		}

		if mouseOnText {
			framesCounter++
		} else {
			framesCounter = 0
		}

		// FETCH WEATHER DATA
		if rl.IsKeyPressed(rl.KeyEnter) && inputText != "" && time.Since(lastFetchTime) > fetchCooldown {
			statusMessage = "Fetching..."
			statusColor = rl.Blue
			fetchedWeather, err := fetchWeatherData(inputText)
			if err == nil {
				weather = fetchedWeather
				statusMessage = "Data fetched successfully!"
				statusColor = rl.Green
				lastFetchTime = time.Now()
			} else {
				statusMessage = fmt.Sprintf("Error: %v", err)
				statusColor = rl.Red
			}
			statusClearTime = time.Now().Add(3 * time.Second)
		}

		if statusMessage != "" && time.Now().After(statusClearTime) {
			statusMessage = ""
		}

		// BEGIN DRAW
		rl.BeginDrawing()

		rl.ClearBackground(rl.RayWhite)

		rl.DrawTextEx(
			font,
			"PLACE MOUSE OVER INPUT BOX!",
			rl.NewVector2(280, 50), 20, 0, rl.Gray,
		)

		rl.DrawRectangleRec(textBox, rl.LightGray)

		if mouseOnText {
			rl.DrawRectangleLines(
				int32(textBox.X),
				int32(textBox.Y),
				int32(textBox.Width),
				int32(textBox.Height),
				rl.Red,
			)
		} else {
			rl.DrawRectangleLines(
				int32(textBox.X),
				int32(textBox.Y),
				int32(textBox.Width),
				int32(textBox.Height),
				rl.DarkGray,
			)
		}

		// CONVERT RUNES TO STRING BEFORE DRAWING
		inputText = string(name[:letterCount])
		rl.DrawTextEx(
			font,
			inputText,
			rl.NewVector2(textBox.X+5, textBox.Y+8), 40, 0, rl.Maroon,
		)

		rl.DrawTextEx(
			font,
			fmt.Sprintf("INPUT CHARS: %d/%d", letterCount, MAX_INPUT_CHARS),
			rl.NewVector2(315, 155), 20, 0, rl.DarkGray,
		)

		rl.DrawTextEx(
			font,
			fmt.Sprintf("INPUT TEXT: %s", inputText),
			rl.NewVector2(315, 180), 20, 0, rl.DarkGray,
		)

		if statusMessage != "" {
			rl.DrawTextEx(
				font,
				statusMessage,
				rl.NewVector2(315, 200), 16, 0, statusColor,
			)
		}

		rl.DrawTextEx(
			font,
			"Press ENTER to fetch weather",
			rl.NewVector2(270, 135), 16, 0, rl.DarkGray,
		)

		if mouseOnText {

			if letterCount < MAX_INPUT_CHARS {

				// DRAW BLINKING UNDERSCORE CHAR_
				if ((framesCounter / 20) % 2) == 0 {
					textWidth := rl.MeasureTextEx(font, inputText, 40, 0).X

					rl.DrawTextEx(
						font,
						"_",
						rl.NewVector2(textBox.X+8+textWidth, textBox.Y+12), 40, 0, rl.Maroon,
					)
				}

			}

			// else {
			// 	rl.DrawTextEx(
			// 		font,
			// 		"Press BACKSPACE to delete chars...",
			// 		rl.NewVector2(230, 180), 20, 0, rl.Gray,
			// 	)
			// }
		}

		// DRAW WEATHER UI
		// if no weather data
		if weather.Location == "" {
			rl.DrawTextEx(
				font,
				"No weather data available",
				rl.NewVector2(270, 240), 20, 0, rl.DarkGray,
			)
		} else {

			weatherBox := rl.NewRectangle(50, 220, 700, 200)
			rl.DrawRectangleRec(weatherBox, rl.NewColor(240, 240, 240, 255))
			rl.DrawRectangleLinesEx(weatherBox, 2, rl.DarkGray)

			rl.DrawTextEx(
				font,
				weather.Location,
				rl.NewVector2(70, 240), 32, 0, rl.DarkBlue,
			)

			rl.DrawTextEx(
				font,
				fmt.Sprintf("%d°C", weather.Temperature),
				rl.NewVector2(70, 280), 48, 0, rl.Black,
			)

			rl.DrawTextEx(
				font,
				weather.Condition,
				rl.NewVector2(200, 290), 24, 0, rl.DarkGray,
			)

			rl.DrawTextEx(
				font,
				fmt.Sprintf("Feels like: %d°C", weather.FeelsLike),
				rl.NewVector2(70, 340), 18, 0, rl.Gray,
			)

			rl.DrawTextEx(
				font,
				fmt.Sprintf("Humidity: %d%%", weather.Humidity),
				rl.NewVector2(400, 240), 20, 0, rl.DarkGray,
			)

			rl.DrawTextEx(
				font,
				fmt.Sprintf("Wind: %.1f km/h", weather.WindSpeed),
				rl.NewVector2(400, 270), 20, 0, rl.DarkGray,
			)
		}

		rl.EndDrawing()
	}
}
