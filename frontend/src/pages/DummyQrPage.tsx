
import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom'; 
import { Button } from '../components/ui/button';
import LoadingSpinner from '../components/common/LoadingSpinner';
import Alert from '../components/common/Alert';
import { formatCurrency } from '../utils/helpers';
import { useQuery } from '@tanstack/react-query';
import { getBookingById } from '../api/booking';
import { QRCodeCanvas } from 'qrcode.react';

const DummyQrPage: React.FC = () => {
  const { bookingId } = useParams<{ bookingId: string }>();
  const navigate = useNavigate();


  const { data: booking, isLoading: isLoadingBooking, isError: isErrorBooking, error: errorBooking } = useQuery({
    queryKey: ['bookingDetails', bookingId],
    queryFn: () => getBookingById(bookingId!),
    enabled: !!bookingId,
    staleTime: Infinity, 
    refetchInterval: (query) => (query.state.data && query.state.data.status === 'pending' ? 5000 : false), 
  });

  const [timeLeft, setTimeLeft] = useState({ total: 0, minutes: 0, seconds: 0 });
  const [paymentStatus, setPaymentStatus] = useState<'pending' | 'success' | 'expired'>('pending');

  useEffect(() => {
    let timer: NodeJS.Timeout;

    const updateCountdown = () => {
      if (!booking || !booking.expires_at) {
        setTimeLeft({ total: 0, minutes: 0, seconds: 0 });
        setPaymentStatus('expired');
        return;
      }

      const expiresTime = new Date(booking.expires_at).getTime();
      const difference = expiresTime - Date.now();

      if (difference <= 0) {
        setTimeLeft({ total: 0, minutes: 0, seconds: 0 });
        setPaymentStatus('expired');


        return;
      }

      setTimeLeft({
        total: difference,
        minutes: Math.floor((difference / 1000 / 60) % 60),
        seconds: Math.floor((difference / 1000) % 60),
      });

      timer = setTimeout(updateCountdown, 1000);
    };

    if (!isLoadingBooking && booking) {
      if (booking.status === 'confirmed') {
        setPaymentStatus('success');
      } else if (booking.status === 'cancelled' || booking.status === 'failed') {
        setPaymentStatus('expired'); 
      } else { 
        setPaymentStatus('pending');
        updateCountdown(); 
      }
    }

    return () => clearTimeout(timer);
  }, [isLoadingBooking, booking, bookingId]);


  if (isLoadingBooking) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <LoadingSpinner />
      </div>
    );
  }

  if (isErrorBooking || !booking) {
    return <Alert type="error" description={`Failed to load booking details: ${errorBooking?.message || 'Booking not found.'}`} />;
  }



  const qrCodeContent = `http://dummy-payment-gateway.com/pay?booking_id=${booking.id}&amount=${booking.total_price}`;

  return (
    <div className="flex flex-col items-center justify-center min-h-screen bg-background p-4 text-center"> {}
      <h1 className="text-3xl font-bold text-foreground mb-4">Complete Your E-Wallet Payment</h1> {}
      <p className="text-xl text-foreground mb-2">Booking ID: {booking.id}</p> {}
      <p className="text-xl text-foreground mb-6">Total Amount: <span className="font-bold text-primary">{formatCurrency(booking.total_price)}</span></p> {}

      {paymentStatus === 'pending' && (
        <>
          <Alert type="info" description="Scan this QR code with your preferred E-Wallet app to complete the payment." className="mb-6 max-w-md" />
          <div className="bg-card p-6 rounded-lg shadow-md mb-6"> {}
            <QRCodeCanvas 
              value={qrCodeContent} 
              size={256} 
              level="H" 
              includeMargin={false} 
              className="w-64 h-64 mx-auto border border-border rounded-md" 
            />
            <p className="mt-4 text-lg font-semibold text-foreground"> {}
              Time Remaining: {timeLeft.minutes < 10 ? '0' + timeLeft.minutes : timeLeft.minutes}:
              {timeLeft.seconds < 10 ? '0' + timeLeft.seconds : timeLeft.seconds}
            </p>
          </div>
          <Button 
            onClick={() => { navigate(`/bookings/${booking.id}`); }} 
            className="bg-secondary text-secondary-foreground hover:bg-secondary/80 mb-4" 
          >
            I Have Paid (Check Status)
          </Button>
        </>
      )}

      {paymentStatus === 'success' && (
        <Alert type="success" title="Payment Successful!" description="Your booking has been confirmed." className="mb-6 max-w-md" />
      )}

      {paymentStatus === 'expired' && (
        <Alert type="error" title="Time Expired!" description="The payment window has closed. Your booking might be cancelled or failed." className="mb-6 max-w-md" />
      )}
      
      <Button 
        onClick={() => navigate('/my-bookings')} 
        className="bg-primary hover:bg-primary/90 text-primary-foreground" 
      >
        Go to My Bookings
      </Button>
    </div>
  );
};

export default DummyQrPage;