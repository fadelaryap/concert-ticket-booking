
import React, { useState, useEffect, type JSX } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { getMyBookings, cancelBooking } from '../api/booking';
import LoadingSpinner from '../components/common/LoadingSpinner';
import Alert from '../components/common/Alert';
import { formatDate, formatCurrency } from '../utils/helpers';
import { Link } from 'react-router-dom';

import { Button } from '../components/ui/button';
import type { Booking } from '../api/booking';

function MyBookingsPage() {
  const queryClient = useQueryClient();
  console.log("booking-page");

  const { data: bookings, isLoading, isError, error } = useQuery<Booking[]>({
    queryKey: ['bookings', 'my'],
    queryFn: getMyBookings,
    staleTime: 60 * 1000,
    refetchInterval: (data) => (data && Array.isArray(data) && data.some(b => b.status === 'pending') ? 5000 : false),
  });

  const cancelBookingMutation = useMutation({
    mutationFn: cancelBooking,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['bookings', 'my'] });
      queryClient.invalidateQueries({ queryKey: ['concerts'] });
      alert('Booking cancelled successfully!');
    },
    onError: (err: any) => {
      alert(`Cancellation failed: ${err.response?.data?.error || err.message}`);
      console.error('Cancellation error:', err);
    },
  });

  const handleCancel = (bookingId: number) => {
    if (window.confirm('Are you sure you want to cancel this booking? This action cannot be undone.')) {
      cancelBookingMutation.mutate(String(bookingId));
    }
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <LoadingSpinner />
      </div>
    );
  }

  if (isError) {
    return <Alert type="error" description={`Failed to load your bookings: ${error.message}`} />;
  }

  if (!bookings) {
    return (
      <div className="w-full p-4">
        <h2 className="text-2xl font-bold mb-4">My Bookings</h2>
        <Alert type="info" description="No booking data found." />
      </div>
    );
  }

  return (
    <div className="w-full p-4"> 
      <h2 className="text-2xl font-bold mb-4">My Bookings</h2>
      {bookings.length > 0 ? (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {bookings.map((booking) => (
            <div key={booking.id} className="bg-white rounded-lg shadow-md p-6 flex flex-col h-full">
              <h3 className="text-xl font-semibold mb-2 text-blue-700">Booking ID: {booking.id}</h3>
              <p className="text-gray-700 text-sm mb-1"><strong>Concert:</strong> {booking.concert_name}</p>
              <p className="text-gray-700 text-sm mb-1"><strong>Date:</strong> {formatDate(booking.concert_date)}</p>
              <p className="text-gray-700 text-sm mb-1"><strong>Total Price:</strong> {formatCurrency(booking.total_price)}</p>
              <p className="text-gray-700 text-sm mb-4"><strong>Status:</strong> {booking.status}</p>
              
              {booking.status === 'pending' && booking.expires_at && (
                <CountdownTimer expiresAt={booking.expires_at} bookingId={booking.id} onExpired={() => queryClient.invalidateQueries({ queryKey: ['bookings', 'my'] })} />
              )}
              {booking.status === 'pending' && (!booking.expires_at || new Date(booking.expires_at).getTime() < new Date().getTime()) && (
                <Alert type="error" description="This booking has expired or is pending payment." />
              )}


              <div className="mt-auto flex space-x-2">
                <Link to={`/bookings/${booking.id}`}>
                  <Button variant="outline" className="bg-blue-100 text-blue-700 hover:bg-blue-200">View Details</Button>
                </Link>
                {booking.status === 'pending' && (
                  <Link to={`/payment/${booking.id}`}>
                    <Button className="bg-green-600 hover:bg-green-700 text-white">Pay Now</Button>
                  </Link>
                )}
                {booking.status === 'pending' && (
                  <Button
                    variant="outline"
                    onClick={() => handleCancel(booking.id)}
                    disabled={cancelBookingMutation.isPending}
                    className="bg-red-100 text-red-700 hover:bg-red-200"
                  >
                    {cancelBookingMutation.isPending ? <LoadingSpinner size="small" /> : 'Cancel'}
                  </Button>
                )}
              </div>
            </div>
          ))}
        </div>
      ) : (
        <Alert type="info" description="You have no bookings yet. Start by booking a concert!" />
      )}
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

export default MyBookingsPage;