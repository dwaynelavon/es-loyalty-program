# Happy Path Data Flow

1. `userCreate` Mutation is ran and the GQL resolver dispatched the `CreateUser` command with email and username

2. `user.commandHandler` handles the `CreateUser` command and invokes `user.Create`. The `user.Create` method will apply the changes to the aggregate root. The aggregate root should be then have 1

3. The dispatcher publishes the events to the event observable

-   The read model handles the event message
-   Side effects like `email` handlers handle the event message
