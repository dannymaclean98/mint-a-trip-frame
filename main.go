package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	// Parse templates
	templatesDir := "template"
	template := template.Must(template.ParseFiles(filepath.Join(templatesDir, "template.html")))

	// Handle requests on the root path
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Set content type to HTML
		w.Header().Set("Content-Type", "text/html")

		randomMemeUrl := GetRandomMeme()

		if randomMemeUrl == "Error" {
			fmt.Println("Error getting meme")
			return
		}

		switch r.Method {
		case "GET":
			if err := template.Execute(w, randomMemeUrl); err != nil {
				http.Error(w, "Failed to render template", http.StatusInternalServerError)
			}
		case "POST":
			if err := template.Execute(w, randomMemeUrl); err != nil {
				http.Error(w, "Failed to render template", http.StatusInternalServerError)
			}
		default:
			// Respond with a 405 Method Not Allowed if the method is not GET or POST
			w.WriteHeader(http.StatusMethodNotAllowed)
			fmt.Fprintf(w, "<h1>405 Method Not Allowed</h1>")
		}
	})

	// Start the HTTP server on port 8080 and handle errors
	fmt.Println("Server listening on port", port)
	httpError := http.ListenAndServe(":"+port, nil)
	if httpError != nil {
		fmt.Println("Error starting server: ", httpError)
		return
	}
}

type Embed struct {
	URL string `json:"url"`
}

type CastBody struct {
	EmbedsDeprecated  []string `json:"embedsDeprecated"`
	Mentions          []int64  `json:"mentions"`
	ParentUrl         string   `json:"parentUrl"`
	Text              string   `json:"text"`
	MentionsPositions []int16  `json:"mentionsPositions"`
	Embeds            []Embed  `json:"embeds"`
}

type Data struct {
	Type        string   `json:"type"`
	FID         int64    `json:"fid"`
	Timestamp   int64    `json:"timestamp"`
	Network     string   `json:"network"`
	CastAddBody CastBody `json:"castAddBody"`
}

type Messages struct {
	Data            Data   `json:"data"`
	Hash            string `json:"hash"`
	HashScheme      string `json:"hashScheme"`
	SignatureScheme string `json:"signatureScheme"`
	Signer          string `json:"signer"`
}

type MemeResponse struct {
	Messages      []Messages `json:"messages"`
	NextPageToken string     `json:"nextPageToken"`
}

func GetRandomMeme() string {
	url := "https://hub.pinata.cloud/v1/castsByParent?url=chain://eip155:1/erc721:0xfd8427165df67df6d7fd689ae67c8ebf56d9ca61"

	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error making the request: %v\n", err)
		return "Error"
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return "Error"
	}

	var apiResponse MemeResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		fmt.Printf("Error parsing JSON response: %v\n", err)
		return "Error"
	}

	var matchingURLs []string

	for _, message := range apiResponse.Messages {
		for _, embed := range message.Data.CastAddBody.Embeds {
			if strings.HasSuffix(embed.URL, ".png") || strings.HasSuffix(embed.URL, ".jpg") || strings.HasSuffix(embed.URL, ".gif") {
				fmt.Println("Found matching URL:", embed.URL)
				matchingURLs = append(matchingURLs, embed.URL)
			}
		}
	}

	if len(matchingURLs) == 0 {
		fmt.Println("No matching URLs found")
		return "Error"
	}

	randomIndex := rand.Intn(len(matchingURLs))

	selectedURL := matchingURLs[randomIndex]

	fmt.Printf("Selected meme: %+v\n", selectedURL)
	return selectedURL
}
