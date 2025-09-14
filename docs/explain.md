### Key Issues & Clarifying Questions for the New Design

This new design is much stronger, but it introduces new considerations, mainly around group dynamics and session lifecycle.

1.  **Group Ordering & Bill Aggregation**:
    *   **Scenario**: A group of 4 people sits at a table. Each person logs in with their own phone number and PIN.
    *   **Question**: How does the kitchen and waitstaff see the *entire order for the table*? When a chef prepares a dish, they need to know it goes to Table 5, not just to "User X".
    *   **Proposed Solution**: The backend can easily handle this. When the kitchen view requests active orders, the query will group `Order`s by `TableIDAtTimeOfOrder`. The UI can display "Table 5" and then list all items for that table, perhaps with a sub-note of which user ordered it (e.g., "Shrimp Tempura - User: ...1234").
    *   **Question for You**: Is it important for one person at the table to see what their friends have ordered on their own device, or is each person's ordering experience completely isolated to their own phone?
    *   **Answer**: I that would be an isolated infomation for each persons in a table.

2.  **Table Status Management**:
    *   With the old design, a table's status was explicitly tied to the `TableSession`. Now, a table's status is *derived*. A table is `Occupied` if `COUNT(UserSession) WHERE CurrentTableID = X > 0`.
    *   **Question**: What is the exact trigger for a table becoming `Free`? Is it when the *last* person associated with that table either logs out, pays, or attaches to a new table? This requires a check every time a user's session status or location changes.
    *   **Answer**: To trigger status table, will be trigger by manual if it was a buffet, but if is Ordering table, it will free when payment successfull. But i think that could be the role of waiter of handle free table, because logic of handle free table in buffet restaurant is hard. But their will be have an option to trigger free all table when end of day or somethings like that.

3.  **Session Lifecycle and Timeouts**:
    *   **Scenario**: A customer logs in, sits at a table, but then closes their browser or their phone dies without logging out. Their `UserSession` is still `Active` and their `CurrentTableID` is still linked, making the table appear `Occupied` forever.
    *   **Question**: How should the system handle these "abandoned" sessions?
    *   **Proposed Solution**: We should implement a session timeout. If no activity (no API calls) is received from a `UserSession` for a certain period (e.g., 2 hours), a background job or scheduled task should automatically mark the session as `Cancelled` or `Abandoned` and free up the table.
    * **Answer**: this could be a great feature, I think restaurant should have a time limit for buffet account with about default is 3 hours and after that time the account could be `EndOfSession`. They will not be able to order new food.

4.  **Account Persistence (One-Time vs. Long-Term)**:
    *   **Question**: Is the `UserAccount` created for a single visit (temporary), or is it a persistent account that the user can use next time they visit?
    *   **Implication**: If it's persistent, the system can offer features like order history. If it's temporary, the `UserAccount` and `UserSession` can be cleaned up/archived after the visit. The design supports both, but it's a key business decision to make.
    * **Answer**: That could make user experiences better, because they can tracking waht they was eat in the past and can try new food or get the food they liked. So, history of order feature is the good things.

After anwser these question abnout, did you have any more issue with the system design please let me know, if not, make a full clearly system design about this project before handle on to working.