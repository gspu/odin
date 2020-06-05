package api

import (
    "context"
    "io/ioutil"
    "net/http"
    "strings"

    "github.com/go-chi/chi"
    "github.com/theycallmemac/odin/odin-engine/pkg/jobs"

    "go.mongodb.org/mongo-driver/mongo"
)

// create resource type to be used by the router
type statsResource struct{}

// create JobStats type to be used for accessing and storing job stats information
type JobStats struct {
    ID string
    Description string
    Type string
    Value string
    Timestamp string
}

func (rs statsResource) Routes() chi.Router {
    // establish new chi router
    r := chi.NewRouter()

    // define routes under the schedule endpoint
    r.Post("/add", rs.AddJobStats)
    r.Post("/get", rs.GetJobStats)
    return r
}

func (rs statsResource) AddJobStats(w http.ResponseWriter, r *http.Request) {
    d, _ := ioutil.ReadAll(r.Body)
    args := strings.Split(string(d), ",")
    typeOfValue, desc, value, id, timestamp := args[0], args[1], args[2], args[3], args[4]
    client, err := jobs.SetupClient()
    if err != nil {
        w.Write([]byte("MongoDB cannot be accessed at the moment\n"))
    } else {
	if InsertIntoMongo(client, typeOfValue, desc, value, id, timestamp) {
            w.Write([]byte("200"))
        } else{
            w.Write([]byte("500"))
        }
    }
}

func InsertIntoMongo(client *mongo.Client, typeOfValue string, desc string, value string, id string, timestamp string) bool {
    var js JobStats
    js.ID = id
    js.Description = desc
    js.Type = typeOfValue
    js.Value = value
    js.Timestamp = timestamp
    collection := client.Database("odin").Collection("observability")
    _, err := collection.InsertOne(context.TODO(), js)
    client.Disconnect(context.TODO())
    if err != nil {
        return false
    }
    return true
}

func (rs statsResource) GetJobStats(w http.ResponseWriter, r *http.Request) {
    d, _ := ioutil.ReadAll(r.Body)
    client, err := jobs.SetupClient()
    if err != nil {
        w.Write([]byte("MongoDB cannot be accessed at the moment\n"))
    } else {
        statsList := jobs.GetJobStats(client, string(d))
        w.Write([]byte(jobs.Format("ID", "DESCRIPTION", "TYPE", "VALUE")))
        for _, stat := range statsList {
            w.Write([]byte(jobs.Format(stat.ID, stat.Description, stat.Type, stat.Value)))
        }
    }
}
