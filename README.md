# Concert Ticket Booking System

This project is a concert ticket booking system. It uses a microservices architecture for different parts of the system.

## Project Structure

This system has three main parts (microservices):

* **User Service**: Manages user accounts, login, and profile.
* **Booking Service**: Handles concert information, seat management, and ticket booking.
* **Payment Service**: Processes payments for bookings.

## Technologies Used

* **Backend (Go)**:
    * Gin Gonic (Web Framework)
    * GORM (ORM for Database)
    * MySQL (Database)
    * Redis (Caching and Rate Limiting)
    * RabbitMQ (Message Broker for asynchronous tasks)
    * JWT (for Authentication)
    * Bcrypt (for Password Hashing)
* **Frontend (React)**:
    * React.js (Frontend Library)
    * Vite (Build Tool)
    * Tailwind CSS (Styling)
    * React Query (Data Fetching)
    * Formik & Yup (Form Handling and Validation)
    * Shadcn UI (UI Components)
    * QRCode.react (QR Code Generation)
* **Containerization**:
    * Docker Compose

## How to Set Up and Run

1.  **Prerequisites**: Make sure you have Docker and Docker Compose installed on your machine.

2.  **Clone the Repository**:
    ```bash
    git clone https://github.com/fadelaryap/concert-ticket-booking
    cd concert-booking
    ```

3.  **Build and Run Services**:
    This command will build all Docker images and start all services (MySQL, Redis, RabbitMQ, User Service, Booking Service, Payment Service, Frontend).
    ```bash
    docker compose up --build
    ```
    * **Important**: If you have run `docker compose up` before and made changes to `init_db.sql` or backend code, you might need to clean old Docker volumes to ensure the database is reset with new data.
        ```bash
        docker compose down -v
        docker compose up --build
        ```

4.  **Access the Application**:
    * Frontend: Open your web browser and go to `http://localhost:3000`
    * Swagger UI (User Service): `http://localhost:8080/swagger/index.html`
    * Swagger UI (Booking Service): `http://localhost:8081/swagger/index.html`
    * Swagger UI (Payment Service): `http://localhost:8082/swagger/index.html`
    * RabbitMQ Management UI: `http://localhost:15672` (Login with `guest`/`guest`)

## First Time Login

When you run the application for the first time, you need to **create a new user account** on the login/register page.

## Key Features

* User Registration and Login
* Concert Listing and Details
* Ticket Booking with Quantity Selection
* Buyer and Ticket Holder Information Capture
* Payment Simulation (always successful for demo)
* Booking Management (View, Cancel)
* Automatic Booking Cancellation for expired pending bookings
* Responsive Frontend Design
