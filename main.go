package main

import (
    "database/sql"
    "fmt"
    "html/template"
    "log"
    "net/http"
    "strconv"

    _ "github.com/mattn/go-sqlite3"
)

// Prompt represents a writing prompt entry
type Prompt struct {
    ID        int
    Prompt    string
    Timestamp string
}

type Words struct {
    ID        int
    WordCount int
    Timestamp string // or time.Time if you're using the time package
}

// Initialize the database and create the prompts table if it doesn't exist
func initDB(db *sql.DB) error {
    createTableSQL := `CREATE TABLE IF NOT EXISTS prompts (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        prompt TEXT NOT NULL,
        timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
        completed BOOLEAN NOT NULL DEFAULT 0
    );`
    _, err := db.Exec(createTableSQL)
    return err
}

// Add a new prompt to the database
func addPrompt(db *sql.DB, promptText string) error {
    _, err := db.Exec("INSERT INTO prompts (prompt) VALUES (?)", promptText)
    return err
}

// Fetch uncompleted prompts from the database
func getUncompletedPrompts(db *sql.DB) ([]Prompt, error) {
    rows, err := db.Query("SELECT id, prompt, timestamp FROM prompts WHERE completed = 0")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var prompts []Prompt
    for rows.Next() {
        var prompt Prompt
        if err := rows.Scan(&prompt.ID, &prompt.Prompt, &prompt.Timestamp); err != nil {
            return nil, err
        }
        prompts = append(prompts, prompt)
    }

    return prompts, nil
}

// Mark a prompt as completed
func completePrompt(db *sql.DB, id int) error {
    _, err := db.Exec("UPDATE prompts SET completed = 1 WHERE id = ?", id)
    return err
}

// Handler for adding a new prompt
func addPromptHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
    if r.Method == http.MethodPost {
        prompt := r.FormValue("prompt")
        if err := addPrompt(db, prompt); err != nil {
            http.Error(w, "Could not save prompt: "+err.Error(), http.StatusInternalServerError)
            return
        }
        http.Redirect(w, r, "/view", http.StatusSeeOther)
        return
    }

    tmpl, err := template.ParseFiles("templates/add_prompt.html")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
    tmpl.Execute(w, nil)
}

// Handler to display prompts
func viewPromptsHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
    prompts, err := getUncompletedPrompts(db)
    if err != nil {
        http.Error(w, "Could not retrieve prompts: "+err.Error(), http.StatusInternalServerError)
        return
    }

    tmpl, err := template.ParseFiles("templates/view_prompts.html")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    tmpl.Execute(w, prompts)
}

// Handler to complete a prompt
func completePromptHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
    idStr := r.URL.Query().Get("id")
    if idStr == "" {
        http.Error(w, "Missing ID", http.StatusBadRequest)
        return
    }

    id, err := strconv.Atoi(idStr)
    if err != nil {
        http.Error(w, "Invalid ID", http.StatusBadRequest)
        return
    }

    if err := completePrompt(db, id); err != nil {
        http.Error(w, "Could not mark prompt as completed: "+err.Error(), http.StatusInternalServerError)
        return
    }

    http.Redirect(w, r, "/view", http.StatusSeeOther)
}

// Fetch completed prompts from the database
func getCompletedPrompts(db *sql.DB) ([]Prompt, error) {
    rows, err := db.Query("SELECT id, prompt, timestamp FROM prompts WHERE completed = 1")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var prompts []Prompt
    for rows.Next() {
        var prompt Prompt
        if err := rows.Scan(&prompt.ID, &prompt.Prompt, &prompt.Timestamp); err != nil {
            return nil, err
        }
        prompts = append(prompts, prompt)
    }

    return prompts, nil
}

// Handler to display completed prompts
func viewCompletedPromptsHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
    prompts, err := getCompletedPrompts(db)
    if err != nil {
        http.Error(w, "Could not retrieve completed prompts: "+err.Error(), http.StatusInternalServerError)
        return
    }

    tmpl, err := template.ParseFiles("templates/view_completed_prompts.html")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    tmpl.Execute(w, prompts)
}

// Fetch word count from the database
func getWordCount(db *sql.DB) ([]Words, error) {
    rows, err := db.Query("SELECT id, wordcount, timestamp FROM WritingSession")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var wordscount []Words
    for rows.Next() {
        var words Words
        if err := rows.Scan(&words.ID, &words.WordCount, &words.Timestamp); err != nil {
            return nil, err
        }
        wordscount = append(wordscount, words)
    }

    return wordscount, nil
}

// Add word count to db
func addWordCount(db *sql.DB, wordCount int) error {
    _, err := db.Exec("INSERT INTO WritingSession (WordCount) VALUES (?)", wordCount)
    return err
}

// Handler for adding word count
func addWordCountHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
    if r.Method == http.MethodPost {
        wordcountstr := r.FormValue("wordcount")
	wordcount, err := strconv.Atoi(wordcountstr)
        if err != nil {
            http.Error(w, "Invalid word count: "+err.Error(), http.StatusBadRequest)
            return
        }
        if err := addWordCount(db, wordcount); err != nil {
            http.Error(w, "Could not save wordcount: "+err.Error(), http.StatusInternalServerError)
            return
        }
        http.Redirect(w, r, "/words", http.StatusSeeOther)
        return
    }

    tmpl, err := template.ParseFiles("templates/word_count.html")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
    tmpl.Execute(w, nil)
}

// Add word count for the day
func viewWordCountHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
    wordscount, err := getWordCount(db)
    if err != nil {
        http.Error(w, "Could not retrieve word count: "+err.Error(), http.StatusInternalServerError)
        return
    }

    tmpl, err := template.ParseFiles("templates/word_count.html")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    tmpl.Execute(w, wordscount)
}

func main() {
    // Set up the database
    db, err := sql.Open("sqlite3", "./prompts.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    // Initialize the database
    if err := initDB(db); err != nil {
        log.Fatal(err)
    }

    // Set up HTTP handlers
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        addPromptHandler(w, r, db)
    })
    http.HandleFunc("/add", func(w http.ResponseWriter, r *http.Request) {
        addPromptHandler(w, r, db)
    })
    http.HandleFunc("/view", func(w http.ResponseWriter, r *http.Request) {
        viewPromptsHandler(w, r, db)
    })
    http.HandleFunc("/complete", func(w http.ResponseWriter, r *http.Request) {
        completePromptHandler(w, r, db)
    })
    http.HandleFunc("/view_completed", func(w http.ResponseWriter, r *http.Request) {
        viewCompletedPromptsHandler(w, r, db)
    })
    http.HandleFunc("/words", func(w http.ResponseWriter, r *http.Request) {
	viewWordCountHandler(w, r, db)
    })
    http.HandleFunc("/words/add", func(w http.ResponseWriter, r *http.Request) {
	addWordCountHandler(w, r, db)
    })

    // Start the web server
    fmt.Println("Starting server on :7000...")
    if err := http.ListenAndServe(":7000", nil); err != nil {
        log.Fatal("ListenAndServe:", err)
    }
}
