# Doctor's Appointment
- Savannah Informatics Backend Developer Take-Home Assessment
- The system is implemented in *Go* using the *Gin Gonic* web framework and *PostgreSQL* as the persistent datastore.
- The architecture is designed to be highly maintainable, prevent booking conflicts (race conditions), and support real-world clinic constraints such as irregular shift structures and midday breaks.

## System Design

### Entity Relationship Diagram (ERD)
- To represent a highly realistic clinic setup without over-complicating storage requirements, we distinguish between standard working hour frames (shifts) and actual bookings. (instead of storing slots, we dynamically generate on-the-fly)
- *Working Hours (Shift)*: Represents active, non-overlapping daily working shifts for a doctor on specific days of the week. By allowing multiple rows per day of the week, this entity seamlessly accomodates midday breaks.
- *Appointment*: Represents a 30-minute booking reservation linking a `Patient` and a `Doctor` on a specific calendar date.

[![alt text](<savannah informatics OA - doctor appointment-1.png>)](https://dbdiagram.io/d/savannah-informatics-OA-doctor-appointment-6a85bacd6440800f52a25235)

## Author
[waltertaya](https://waltertaya.pages.dev/)
