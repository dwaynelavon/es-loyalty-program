# Event Sourced Loyalty Program

POC for a loyalty program with event sourcing in Go.

## Business Logic

-   Users get 100 points for signing up without a referral
-   Users get 200 points for referring a user
-   Users get 200 points for signing up with a referral code
-   Users get 50 points for creating their profile

## Events

### Wallet Aggregate

-   PointsEarned

### User Aggregate

-   UserCreated
-   UserDeleted
-   ReferralCreated
-   ReferralCompleted

### TODO

-   Add Flow chart illustrating data flow
-   Add fx modules
-   Snapshots
-   Pause and Restart projectors
-   Enforce ordering in the read model. Maybe locks updates for a particular aggregateID until processing finishes
-   Event handlers load events into memory, apply the new events on the read model aggregate, then save changes to ensure business logic
-   Consistent comments in code
-   PointsRedeemed
-   How to ensure unique values with eventual consistency (unique username for CreateUser). Maybe the approach is to have a immediately consistent data store that houses all of the usernames in the system. Then, check that datastore and update it before accepting the command to CreateUser

## Ideas

```go
type ProjectorStatus string

const (
	ProjectorStatusStopped = "Stopped"
	ProjectorStatusPaused  = "Active"
)

type Projector interface {
	Status() ProjectorStatus
	Start() error
	Stop() error
}
```
