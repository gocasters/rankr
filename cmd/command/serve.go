package command

import (
	"github.com/spf13/cobra"
	"log"
	"net/http"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the rankr application",
	Long:  `This command starts the rankr app.`,
	Run: func(cmd *cobra.Command, args []string) {
		serve()
	},
}

func serve() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"message":"Hello from Go HTTP server","status":"ok"}`))
	})

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	RootCmd.AddCommand(serveCmd)
}
