package main

//TODO
// - HTTP error messages returned back to the calling program
// - use HTML templates to handle the embedded HTML code - getJobHandler,getJobNoteHandler
// - migrate panics to better error upstream handling
// - implement better & additional form/request data sanitizing and validation
// - move datastore functionality to data.go and replace with abstractions or ORM
import (
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/ostafen/clover"
)

var jobsdb *clover.DB
var docs = make([]*clover.Document, 0)

type JobViewData struct {
	Filter   string
	Sort     int
	Jobitems []Jobs
}

// default index page handler
func indexHandler(w http.ResponseWriter, r *http.Request) {
	index.Render(w, nil)
}

// retrieve a list of jobs that can be filtered or sorted
func getJobsHandler(w http.ResponseWriter, r *http.Request) {

	params := mux.Vars(r)
	// determine the sorting index
	sortcol, err := strconv.Atoi(params["s"])
	_, ok := params["s"]
	if ok && err != nil {
		http.Redirect(w, r, "/jobs", http.StatusFound)
	}

	// filter flags to string mapping
	jobstatus := map[string]string{"s": "scheduled", "a": "active",
		"i": "invoicing", "t": "to priced", "c": "completed"}

	//retrieve a list of jobs, either filtered or as all
	filter := ""
	if len(strings.ToLower(params["f"])) > 0 {
		filter = params["f"]
		js := jobstatus[filter]
		// retrieve the filtered data
		docs, err = jobsdb.Query("jobs").Where(clover.Field("status").Eq(js)).FindAll()
		if err != nil {
			log.Println("Job data not found")
			panic(err.Error())
		}
	} else {
		// retrieve ALL the data
		docs, err = jobsdb.Query("jobs").FindAll()
		if err != nil {
			log.Println("Job data not found")
			panic(err.Error())
		}
	}

	//fit the retrieved data into a datastructure suitable for the template view
	_job := &Jobs{}
	viewdata := JobViewData{}
	viewdata.Filter = filter
	viewdata.Sort = sortcol
	for _, doc := range docs {
		doc.Unmarshal(_job)
		jobitem := Jobs{_job.Jobid, _job.Status, _job.Timestamp, _job.Name,
			_job.Address, _job.Phone, _job.Email, _job.Notes}
		viewdata.Jobitems = append(viewdata.Jobitems, jobitem)
	}
	//sort the view data before sending it back to the template view
	switch sortcol {
	case 1:
		sort.SliceStable(viewdata.Jobitems, func(i, j int) bool {
			return viewdata.Jobitems[i].Name.Last < viewdata.Jobitems[j].Name.Last
		})
	case 2:
		sort.SliceStable(viewdata.Jobitems, func(i, j int) bool {
			return viewdata.Jobitems[i].Status < viewdata.Jobitems[j].Status
		})
	default:
		sort.SliceStable(viewdata.Jobitems, func(i, j int) bool {
			return viewdata.Jobitems[i].Jobid < viewdata.Jobitems[j].Jobid
		})

	}
	//	log.Println(viewdata) //diagnostics
	jobs.Render(w, viewdata)
}

//retrieve the job details for a single job
func getJobHandler(w http.ResponseWriter, r *http.Request) {

	// determine the sorting index
	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		log.Printf("Invalid job number")
		return
	}

	//find the job details
	docs, err = jobsdb.Query("jobs").Where(clover.Field("jobid").Eq(id)).FindAll()
	if err != nil {
		log.Printf("No job details found")
		return
	}

	_job := &Jobs{}
	docs[0].Unmarshal(_job) //there should only be one job per id in the dataset

	//preformat the notes
	notes := ""
	for _, n := range _job.Notes {
		notes += n + "<br />"
	}

	s := `<div class="card border-dark mb-3" style="width: 40rem;">
  		 <div class="card-body text-dark">
    	 <h3 class="card-title">Job Details</h3>
    	 <h5 class="card-subtitle mb-2 text-muted">Job # ` + strconv.Itoa(_job.Jobid) + `</h5>    	     	 
		 <ul class="list-group list-group-flush">
		
    		<li class="list-group-item"><strong>Email:</strong> ` + _job.Email + `</li>
    		<li class="list-group-item"><strong>Address:</strong> ` + _job.Address + `</li>
    		<li class="list-group-item"><strong>Phone:</strong> ` + _job.Phone + `</li>
    		<li class="list-group-item"><strong>Created:</strong> ` + _job.Timestamp + `</li>			
			<li class="list-group-item"><strong>Notes:</strong> <br />` + notes + `</li>
  		 </ul></div></div>`

	w.Write([]byte(s))
}

// This handler should allow for any job field to be updated. currently it only has code to update
// the status flag
func editJobHandler(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()
	if err != nil {
		log.Printf("Error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//check the passed id is a valid number - no range checking done
	_jobid := r.Form.Get("id")
	id, err := strconv.Atoi(_jobid)
	if err != nil {
		log.Printf("Error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// use the filter flag as a mapping to job status
	jobstatus := map[string]string{"s": "scheduled", "a": "active",
		"i": "invoicing", "t": "to priced", "c": "completed"}
	updates := make(map[string]interface{})

	//additional form data can be validated and sanitized here before the update
	status := r.Form.Get("stat")
	_, ok := jobstatus[status]
	//only update if the status code is valid/exists
	if len(strings.ToLower(status)) > 0 && ok {
		//additional fields for updating can be added here
		updates["status"] = jobstatus[status]
		jobsdb.Query("jobs").Where(clover.Field("jobid").Eq(id)).Update(updates)
	}

	http.Redirect(w, r, "/jobs", http.StatusFound)
}

func getJobNoteHandler(w http.ResponseWriter, r *http.Request) {
	//check the passed id is a valid number - no range checking done
	params := mux.Vars(r)
	// determine the sorting index
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		log.Printf("Invalid job number")
		return
	}

	docs, err = jobsdb.Query("jobs").Where(clover.Field("jobid").Eq(id)).FindAll()
	if err != nil {
		log.Printf("No job details found")
		return
	}

	_job := &Jobs{}
	for _, doc := range docs {
		doc.Unmarshal(_job)
	}

	//add a URL/icon to the note to edit/delete
	//this embedded HTML could have been implemented with templates
	notes := ""
	i := 1
	for _, n := range _job.Notes {
		notes += `<tr><td>`
		notes += `<input type="text" class="form-control" name="note` + strconv.Itoa(i) + `" id="note` + strconv.Itoa(i) + `" placeholder="Enter a job note" value="` + n + `"> `
		notes += `<input type="button" class="btn btn-success" onclick="deleteNoteRow(this)" value="X">`
		notes += `</td</tr>`
		i++
	}

	s := `<div class="card border-dark mb-3" style="width: 40rem;">
  		  <div class="card-body text-dark">
    	  <h3 class="card-title">Job Notes</h3>
    	  <h5 class="card-subtitle mb-2 text-muted">Job # ` + strconv.Itoa(_job.Jobid) + `</h5>    	     	 
		  <form id="notesform" class="form-inline"> 
		  <table id="notes-table" class="table table-striped table-dark"> 
			<thead><th scope="col"><strong>Add or Edit Job Notes</strong>&nbsp;&nbsp;&nbsp;&nbsp;  
			<button type="button" class="btn btn-success" onclick="addNoteRow()">Add Note</button>
			<button type="button" class="btn btn-success" onclick="updJobNotes(` + strconv.Itoa(_job.Jobid) + `)">Save Notes</button>
			</th></thead><tbody>` + notes + `</tbody>
		  </table></form>
		  </div></div>`

	w.Write([]byte(s))

}

func editJobNoteHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	// determine the sorting index
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		log.Printf("Invalid job number")
		return
	}

	// parse the incoming form data
	err = r.ParseMultipartForm(100000) // 100KB buffer
	if err != nil {
		log.Printf("Query error: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//retrive the new updated notes and prepare to store them
	newnotes := []string{}
	for _, v := range r.MultipartForm.Value {
		newnotes = append(newnotes, v[0])
	}
	updates := make(map[string]interface{}) //arbitary map holder for the updates
	updates["notes"] = newnotes

	err = jobsdb.Query("jobs").Where(clover.Field("jobid").Eq(id)).Update(updates)
	if err != nil {
		log.Printf("Error updating job notes")
		return
	}

	http.Redirect(w, r, "/jobs", http.StatusFound)
}
