# Event Sourced Loyalty Program

POC for a loyalty program with event sourcing in Go

## Business Logic

-   Users get 100 points for signing up without a referral
-   Users get 200 points for referring a user
-   Users get 200 points for signing up with a referral code
-   Users get 50 points for creating their profile

## Events

### Wallet Aggregate

-   PointsEarned
-   PointsRedeemed

### User Aggregate

-   UserCreated
-   UserDeleted
-   UserProfileCreated
-   ReferralCreated
-   ReferralStatusUpdated (Created, Pending, Rejected, Completed)
-   ReferralCompleted

### TODO

-   Run projections on start up
-   Enforce ordering in the read model.
-   Event handlers load events into memory, apply the new events on the read model aggregate, then save changes
-   Abstract logging into logger service
-   Add config for backOff values
