
import React, { useState, useEffect, type JSX } from 'react';
import { useParams } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { getBookingById, cancelBooking } from '../api/booking';
import LoadingSpinner from '../components/common/LoadingSpinner';
import Alert from '../components/common/Alert';
import { formatDate, formatCurrency } from '../utils/helpers';

import { Button } from '../components/ui/button';
import type { Booking } from '../api/booking';

function BookingDetailPage() {
  const { bookingId } = useParams<{ bookingId: string }>();
  const queryClient = useQueryClient();
  const [cancelError, setCancelError] = useState<string | null>(null);

  const { data: booking, isLoading, isError, error } = useQuery<Booking>({
    queryKey: ['bookingDetails', bookingId],
    queryFn: () => getBookingById(bookingId!),
    enabled: !!bookingId,
    staleTime: 30 * 1000,
    refetchInterval: (query) => {
      const bookingData = query.state.data;
      return bookingData && bookingData.status === 'pending' ? 5000 : false;
    },
  });

  const cancelBookingMutation = useMutation({
    mutationFn: cancelBooking,
    onSuccess: () => {
      setCancelError(null);
      queryClient.invalidateQueries({ queryKey: ['bookings', bookingId] });
      queryClient.invalidateQueries({ queryKey: ['bookings', 'my'] });
      queryClient.invalidateQueries({ queryKey: ['concerts'] });
      alert('Booking cancelled successfully!');
    },
    onError: (err: any) => {
      setCancelError(err.response?.data?.error || err.message || 'Failed to cancel booking.');
      console.error('Cancellation error:', err);
    },
  });

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <LoadingSpinner />
      </div>
    );
  }

  if (isError) {
    return <Alert type="error" description={`Failed to load booking details: ${error.message}`} />;
  }

  if (!booking) {
    return <Alert type="info" description="Booking not found or invalid." />;
  }

  const handleCancel = () => {
    if (window.confirm('Are you sure you want to cancel this booking? This action cannot be undone.')) {
      cancelBookingMutation.mutate(bookingId!);
    }
  };

  const isPending = booking.status === 'pending';
  const hasExpired = booking.expires_at && new Date(booking.expires_at).getTime() < new Date().getTime();

  return (
    <div className="w-full p-4 max-w-2xl mx-auto bg-white rounded-lg shadow-lg"> 
      <h2 className="text-2xl font-bold mb-4 text-blue-700">Booking Details (ID: {booking.id})</h2>
      {cancelError && <Alert type="error" description={cancelError} />}

      <p className="text-gray-700 mb-2"><strong>Concert:</strong> {booking.concert_name}</p>
      <p className="text-gray-700 mb-2"><strong>Date:</strong> {formatDate(booking.concert_date)}</p>
      <p className="text-gray-700 mb-4"><strong>Total Price:</strong> {formatCurrency(booking.total_price)}</p>
      <p className="text-gray-700 mb-4"><strong>Current Status:</strong> {booking.status}</p>
      
      {booking.buyer_info && (
        <div className="mb-4">
          <h3 className="text-lg font-semibold mb-2">Buyer Information:</h3>
          <p className="text-gray-700 text-sm">Name: {booking.buyer_info.full_name}</p>
          <p className="text-gray-700 text-sm">Phone: {booking.buyer_info.phone_number}</p>
          <p className="text-gray-700 text-sm">Email: {booking.buyer_info.email}</p>
          <p className="text-gray-700 text-sm">KTP: {booking.buyer_info.ktp_number}</p>
        </div>
      )}
      {booking.ticket_holder_info && (
        <div className="mb-4">
          <h3 className="text-lg font-semibold mb-2">Ticket Holder Information:</h3>
          <p className="text-gray-700 text-sm">Name: {booking.ticket_holder_info.full_name}</p>
          <p className="text-gray-700 text-sm">KTP: {booking.ticket_holder_info.ktp_number}</p>
        </div>
      )}

      <p className="text-gray-700 mb-4"><strong>Booked Tickets:</strong></p>
      <ul className="list-disc pl-5 mb-4 text-gray-700">
        {booking.booked_seats && booking.booked_seats.length > 0 ? (
          booking.booked_seats.map((seat, index) => (
            <li key={index}>Seat: {seat.seat_number} (Class ID: {seat.ticket_class_id})</li>
          ))
        ) : (
          <li>No tickets associated with this booking.</li>
        )}
      </ul>
      
      {isPending && booking.expires_at && !hasExpired && (
        <CountdownTimer expiresAt={booking.expires_at} bookingId={booking.id} onExpired={() => queryClient.invalidateQueries({ queryKey: ['bookings', booking.id] })} />
      )}
      {isPending && hasExpired && (
        <Alert type="error" description="This booking has expired and will be cancelled soon." />
      )}

      <div className="mt-4 flex space-x-2">
        {isPending && !hasExpired && (
          <Button onClick={handleCancel} disabled={cancelBookingMutation.isPending} className="bg-red-600 hover:bg-red-700 text-white">
            {cancelBookingMutation.isPending ? <LoadingSpinner size="small" /> : 'Cancel Booking'}
          </Button>
        )}
        {booking.status === 'confirmed' && (
          <Alert type="success" description="This booking is confirmed and paid." />
        )}
        {(booking.status === 'cancelled' || booking.status === 'failed') && (
          <Alert type="info" description={`This booking has been ${booking.status}.`} />
        )}
      </div>
    </div>
  );
}

const CountdownTimer: React.FC<{ expiresAt: string; bookingId: number; onExpired: () => void }> = ({ expiresAt, onExpired }) => {
  const calculateTimeLeft = () => {
    const difference = new Date(expiresAt).getTime() - new Date().getTime();
    let timeLeft: { minutes?: number; seconds?: number } = {};

    if (difference > 0) {
      timeLeft = {
        minutes: Math.floor((difference / 1000 / 60) % 60),
        seconds: Math.floor((difference / 1000) % 60),
      };
    }
    return timeLeft;
  };

  const [timeLeft, setTimeLeft] = useState(calculateTimeLeft());

  useEffect(() => {
    const timer = setTimeout(() => {
      setTimeLeft(calculateTimeLeft());
      if (Object.keys(timeLeft).length === 0 && onExpired) {
        onExpired();
      }
    }, 1000);

    return () => clearTimeout(timer);
  }, [timeLeft, expiresAt, onExpired]);

  const timerComponents: JSX.Element[] = [];

  Object.keys(timeLeft).forEach((interval) => {
    const value = (timeLeft as any)[interval];
    if (value !== undefined) {
      timerComponents.push(
        <span key={interval}>
          {value < 10 ? '0' : ''}{value} {interval}{" "}
        </span>
      );
    }
  });

  return (
    <p className="text-yellow-700 bg-yellow-50 border border-yellow-400 p-3 rounded-md mb-4">
      Expires in: {timerComponents.length ? timerComponents : <span>Time's up!</span>}
    </p>
  );
};

export default BookingDetailPage;