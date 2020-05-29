# es-loyalty-program

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
-   RefferalCompleted
