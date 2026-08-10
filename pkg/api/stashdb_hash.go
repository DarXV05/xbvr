package api

import (
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/emicklei/go-restful/v3"
	"github.com/xbapps/xbvr/pkg/config"
	"github.com/xbapps/xbvr/pkg/models"
	"github.com/xbapps/xbvr/pkg/scrape"
)

type StashHashCandidate struct {
	ID        uint   `json:"id"`
	SceneID   string `json:"scene_id"`
	Title     string `json:"title"`
	Site      string `json:"site"`
	SceneURL  string `json:"scene_url"`
	MatchedOn string `json:"matched_on"`
}

type StashHashMatch struct {
	StashID       string               `json:"stash_id"`
	StashURL      string               `json:"stash_url"`
	Title         string               `json:"title"`
	Studio        string               `json:"studio"`
	ReleaseDate   string               `json:"release_date"`
	Duration      int                  `json:"duration"`
	Cover         string               `json:"cover"`
	StudioURLs    []string             `json:"studio_urls"`
	Slug          string               `json:"slug"`
	LinkedSceneID string               `json:"linked_scene_id"`
	Candidates    []StashHashCandidate `json:"candidates"`
}

type StashHashSearchResult struct {
	OsHash  string           `json:"oshash"`
	Message string           `json:"message"`
	Count   int              `json:"count"`
	Matches []StashHashMatch `json:"matches"`
}

var slugTrailingID = regexp.MustCompile(`-\d+$`)

// aggregators and studios publish the same scene under their own trailing id,
// so the slug without it is the only part the two urls share
func sceneSlug(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	seg := path.Base(strings.TrimSuffix(u.Path, "/"))
	if seg == "." || seg == "/" || seg == "" {
		return ""
	}
	return slugTrailingID.ReplaceAllString(seg, "")
}

// stashdb pads os hashes to 16 characters, xbvr does not always
func padOsHash(hash string) string {
	if len(hash) < 16 {
		return strings.Repeat("0", 16-len(hash)) + hash
	}
	return hash
}

func (i ExternalReference) searchStashdbByHash(req *restful.Request, resp *restful.Response) {
	empty := StashHashSearchResult{Matches: []StashHashMatch{}}

	fileID, err := strconv.Atoi(req.PathParameter("file-id"))
	if err != nil {
		empty.Message = "invalid file id"
		resp.WriteHeaderAndEntity(http.StatusBadRequest, empty)
		return
	}

	var file models.File
	if err := file.GetIfExistByPK(uint(fileID)); err != nil || file.ID == 0 {
		empty.Message = "file not found"
		resp.WriteHeaderAndEntity(http.StatusNotFound, empty)
		return
	}

	out := StashHashSearchResult{OsHash: file.OsHash, Matches: []StashHashMatch{}}

	if config.Config.Advanced.StashApiKey == "" {
		out.Message = "No StashDB API key configured"
		resp.WriteHeaderAndEntity(http.StatusOK, out)
		return
	}
	if file.OsHash == "" {
		out.Message = "File has no oshash"
		resp.WriteHeaderAndEntity(http.StatusOK, out)
		return
	}

	queryVariable := `
	{"input":{
		"fingerprints": {
			"value": "` + padOsHash(file.OsHash) + `",
			"modifier": "INCLUDES"
		},
		"page": 1
	}
	}`

	found := scrape.GetScenePage(queryVariable)
	out.Count = found.Data.QueryScenes.Count

	db, _ := models.GetDB()
	defer db.Close()

	for _, stashScene := range found.Data.QueryScenes.Scenes {
		if stashScene.ID == "" {
			continue
		}
		m := StashHashMatch{
			StashID:     stashScene.ID,
			StashURL:    "https://stashdb.org/scenes/" + stashScene.ID,
			Title:       stashScene.Title,
			Studio:      stashScene.Studio.Name,
			ReleaseDate: stashScene.Date,
			Duration:    stashScene.Duration,
			StudioURLs:  []string{},
			Candidates:  []StashHashCandidate{},
		}
		if len(stashScene.Images) > 0 {
			m.Cover = stashScene.Images[0].URL
		}

		var link models.ExternalReferenceLink
		db.Where(&models.ExternalReferenceLink{ExternalSource: "stashdb scene", ExternalId: stashScene.ID}).First(&link)
		if link.ID != 0 {
			m.LinkedSceneID = link.InternalNameId
		}

		seen := map[uint]bool{}
		for _, u := range stashScene.URLs {
			if !strings.EqualFold(u.Type, "STUDIO") {
				continue
			}
			m.StudioURLs = append(m.StudioURLs, u.URL)
			slug := sceneSlug(u.URL)
			if slug == "" {
				continue
			}
			if m.Slug == "" {
				m.Slug = slug
			}
			var scenes []models.Scene
			db.Where("scene_url LIKE ?", "%"+slug+"%").Limit(10).Find(&scenes)
			for _, s := range scenes {
				if seen[s.ID] {
					continue
				}
				seen[s.ID] = true
				m.Candidates = append(m.Candidates, StashHashCandidate{
					ID:        s.ID,
					SceneID:   s.SceneID,
					Title:     s.Title,
					Site:      s.Site,
					SceneURL:  s.SceneURL,
					MatchedOn: slug,
				})
			}
		}

		out.Matches = append(out.Matches, m)
	}

	if out.Count == 0 {
		out.Message = "No StashDB scene carries this hash"
	}

	resp.WriteHeaderAndEntity(http.StatusOK, out)
}
