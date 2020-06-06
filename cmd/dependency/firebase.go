package dependency

import (
	"context"
	"fmt"
	"path"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/dwaynelavon/es-loyalty-program/config"
	"github.com/pkg/errors"
	"google.golang.org/api/option"
)

func NewFirebaseApp() (*firebase.App, error) {
	configReader := config.NewReader()

	firebaseConfigFile, configErr := configReader.ReadFirebaseCredentialsFileLocation()
	if configErr != nil {
		return nil, configErr
	}

	var (
		opt  = &[]option.ClientOption{option.WithCredentialsFile(path.Join("../config", *firebaseConfigFile))}
		conf = &firebase.Config{}
		ctx  = context.Background()
	)

	app, err := firebase.NewApp(ctx, conf, *opt...)
	if err != nil {
		return nil, fmt.Errorf("Error initializing firebase app: %v", err)
	}

	return app, nil
}

func NewFirebaseClient(firebaseApp *firebase.App) (*firestore.Client, error) {
	firestoreClient, errFirestoreClient := firebaseApp.Firestore(context.Background())
	if errFirestoreClient != nil {
		panic(errors.Wrap(errFirestoreClient, "unable to instantiate Firebase"))
	}
	return firestoreClient, nil
}
