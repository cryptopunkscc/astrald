package webdata

import (
	"context"
	"embed"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/sets"
	"github.com/cryptopunkscc/astrald/mod/shares"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/resources"
	"github.com/gin-gonic/gin"
	"html/template"
	"io"
	"net/http"
)

//go:embed assets/* tmpl/*
var res embed.FS

type Module struct {
	config   Config
	node     node.Node
	log      *log.Logger
	assets   resources.Resources
	identity id.Identity
	mux      *http.ServeMux

	storage storage.Module
	shares  shares.Module
	sets    sets.Module
	content content.Module
}

func (mod *Module) Run(ctx context.Context) error {
	gin.DisableConsoleColor()
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.LoggerWithWriter(io.Discard), gin.Recovery())

	templ := template.Must(
		template.New("").
			Funcs(r.FuncMap).
			ParseFS(res, "tmpl/*"))

	r.SetHTMLTemplate(templ)

	// Serve static files from the embedded filesystem
	r.GET("/assets/*filepath", mod.handleAssets)
	r.GET("/sets/:name", mod.handleSetsShow)
	r.GET("/objects/:id/open", mod.handleObjectsOpen)
	r.GET("/objects/:id/show", mod.handleObjectsShow)
	r.GET("/", mod.handleSetsIndex)

	var server = http.Server{
		Addr:    mod.config.Listen,
		Handler: r,
	}

	go server.ListenAndServe()

	<-ctx.Done()
	server.Shutdown(context.Background())

	return nil
}

func (mod *Module) handleAssets(c *gin.Context) {
	// Extract the requested file path from the URL parameter
	filepath := c.Param("filepath")

	// Open the file from the embedded filesystem
	file, err := res.Open("assets" + filepath)
	if err != nil {
		// Handle the error, e.g., file not found
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	defer file.Close()
	var buf = make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.Header("Content-Type", http.DetectContentType(buf))
	c.Header("Content-Disposition", "inline")

	_, err = c.Writer.Write(buf[:n])
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	_, err = io.Copy(c.Writer, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
}
