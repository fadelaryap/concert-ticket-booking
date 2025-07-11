
import api from './index';


export interface ConcertTicketClass {
  id: number;
  concert_id: number;
  name: string;
  price: number;
  total_seats_in_class: number;
  available_seats_in_class: number;
  created_at: string;
  updated_at: string;
}

export interface Concert {
  id: number;
  name: string;
  artist: string;
  date: string; 
  venue: string;
  total_seats: number;
  available_seats: number;
  description: string;
  status: string;
  image_url: string;
  ticket_classes: ConcertTicketClass[]; 
  created_at: string;
  updated_at: string;
}

export interface Seat {
  id: number;
  seat_number: string;
  status: string;
  concert_id: number;
  ticket_class_id: number;
  ticket_class_name?: string; 
}

export interface BookingBuyerInfo {
  id: number;
  full_name: string;
  phone_number: string;
  email: string;
  ktp_number: string;
  booking_id: string; 
}

export interface BookingTicketHolderInfo {
  id: number;
  full_name: string;
  ktp_number: string;
  booking_id: string; 
}

export interface Booking {
  id: string; 
  user_id: number;
  concert_id: number;
  total_price: number;
  status: string;
  payment_id: number | null;
  expires_at: string | null;
  booked_seats: Seat[] | null; 
  concert_name: string;
  concert_date: string;
  buyer_info?: BookingBuyerInfo;
  ticket_holder_info?: BookingTicketHolderInfo;
  created_at: string;
  updated_at: string;
}

export interface TicketQuantityByClassRequest {
  ticket_class_id: number;
  quantity: number;
}

export interface CreateBookingRequest {
  concert_id: number;
  tickets_by_class: TicketQuantityByClassRequest[];
  buyer_info: {
    full_name: string;
    phone_number: string;
    email: string;
    ktp_number: string;
  };
  ticket_holder_info?: {
    full_name: string;
    ktp_number: string;
  } | null;
}

export interface InitiatePaymentRequest {
  booking_id: string; 
  amount: number;
  payment_method: string;
}

export interface PaymentResponse {
  id: number;
  booking_id: string; 
  amount: number;
  payment_method: string;
  transaction_id: string;
  status: string;
  created_at: string;
  updated_at: string;
}



const BOOKING_SERVICE_BASE_PATH = `http://localhost:${import.meta.env.VITE_BOOKING_SERVICE_PORT || 8081}/api/v1`; 
const PAYMENT_SERVICE_BASE_PATH = `http://localhost:${import.meta.env.VITE_PAYMENT_SERVICE_PORT || 8082}/api/v1`;

export const getAllConcerts = async (): Promise<Concert[]> => {
  const response = await api.get(`${BOOKING_SERVICE_BASE_PATH}/concerts`);
  return response.data;
};

export const getConcertById = async (concertId: string): Promise<Concert> => {
  const response = await api.get(`${BOOKING_SERVICE_BASE_PATH}/concerts/${concertId}`);
  return response.data;
};

export const getConcertSeats = async (concertId: string): Promise<Seat[]> => {
  const response = await api.get(`${BOOKING_SERVICE_BASE_PATH}/concerts/${concertId}/seats`);
  return response.data;
};

export const createBooking = async (fullBookingData: CreateBookingRequest): Promise<Booking> => {
  const response = await api.post(`${BOOKING_SERVICE_BASE_PATH}/bookings/`, fullBookingData);
  return response.data;
};

export const getMyBookings = async (): Promise<Booking[]> => {
  const response = await api.get(`${BOOKING_SERVICE_BASE_PATH}/bookings/my`);
  return response.data;
};

export const getBookingById = async (bookingId: string): Promise<Booking> => {
  const response = await api.get(`${BOOKING_SERVICE_BASE_PATH}/bookings/${bookingId}`);
  return response.data;
};

export const cancelBooking = async (bookingId: string): Promise<{ message: string }> => {
  const response = await api.put(`${BOOKING_SERVICE_BASE_PATH}/bookings/${bookingId}/cancel`);
  return response.data;
};

export const initiatePayment = async (paymentData: InitiatePaymentRequest): Promise<PaymentResponse> => {
  const response = await api.post(`${PAYMENT_SERVICE_BASE_PATH}/payments`, paymentData); 
  return response.data;
};

export const getPaymentDetails = async (paymentId: number): Promise<PaymentResponse> => {
  const response = await api.get(`${PAYMENT_SERVICE_BASE_PATH}/payments/${paymentId}`);
  return response.data;
};