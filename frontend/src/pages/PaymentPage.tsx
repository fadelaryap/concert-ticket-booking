
import React, { useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom'; 
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { getBookingById, initiatePayment, getPaymentDetails } from '../api/booking';
import LoadingSpinner from '../components/common/LoadingSpinner';
import Alert from '../components/common/Alert';
import { formatCurrency, getCardType } from '../utils/helpers'; 


import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from '../components/ui/accordion';
import { Button } from '../components/ui/button';
import { Label } from '../components/ui/label';

import { Input } from '../components/ui/input';


interface PaymentInitiationData {
  booking_id: string; 
  amount: number;
  payment_method: string;
}

interface ApiError extends Error {
  response?: {
    data?: {
      error?: string;
    };
  };
}


const ErrorMessage = ({ message }: { message: string }) => (
  <div className="flex items-center justify-center min-h-screen">
    <Alert type="error" description={message} />
  </div>
);

function PaymentPage() {
  const { bookingId } = useParams<{ bookingId: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [selectedPaymentMethod, setSelectedPaymentMethod] = useState<string>(''); 
  const [paymentInitiationError, setPaymentInitiationError] = useState<string | null>(null);
  const [cardNumber, setCardNumber] = useState(''); 
  const [cardType, setCardType] = useState<string | null>(null); 


  const { data: booking, isLoading: isLoadingBooking, isError: isErrorBooking, error: errorBooking } = useQuery({
    queryKey: ['bookingDetails', bookingId],
    queryFn: () => getBookingById(bookingId!),
    enabled: !!bookingId,
    staleTime: 30 * 1000,
    refetchInterval: (query) => (query.state.data && query.state.data.status === 'pending' ? 5000 : false),
  });


  const { data: payment, isLoading: isLoadingPayment } = useQuery({
    queryKey: ['paymentDetails', booking?.payment_id],
    queryFn: () => getPaymentDetails(booking!.payment_id!),
    enabled: !!booking?.payment_id,
    staleTime: 30 * 1000,
  });

  const initiatePaymentMutation = useMutation({
    mutationFn: (data: PaymentInitiationData) => initiatePayment(data),
    onSuccess: (data) => {
      setPaymentInitiationError(null);
      queryClient.invalidateQueries({ queryKey: ['bookingDetails', bookingId] });
      queryClient.invalidateQueries({ queryKey: ['paymentDetails', data.id] });
      queryClient.invalidateQueries({ queryKey: ['bookings', 'my'] });


      if (selectedPaymentMethod === 'e_wallet') {
        navigate(`/qr-payment/${bookingId}`); 
      } else {
        alert('Payment initiated successfully! Please check booking details for status.'); 
      }
    },
    onError: (error: ApiError) => {
      console.error('Payment initiation failed:', error.response?.data?.error || error.message);
      setPaymentInitiationError(error.response?.data?.error || 'Payment failed. Please try again.');
    },
  });


  const handleCardNumberChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value.replace(/\D/g, ''); 
    setCardNumber(value);
    setCardType(getCardType(value)); 
  };

  if (isLoadingBooking || isLoadingPayment) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <LoadingSpinner />
      </div>
    );
  }

  if (isErrorBooking) {
    return <ErrorMessage message={`Failed to load booking details for payment: ${errorBooking?.message}`} />;
  }

  if (!booking) {
    return <ErrorMessage message="Booking not found or invalid." />;
  }

  const isPaid = booking.status === 'confirmed';
  const isFailed = booking.status === 'failed';
  const isCancelled = booking.status === 'cancelled';

  const handlePaymentSubmit = () => {
    if (!selectedPaymentMethod) {
      setPaymentInitiationError('Please select a payment method.');
      return;
    }
    setPaymentInitiationError(null);


    if (selectedPaymentMethod === 'e_wallet') {
      navigate(`/qr-payment/${bookingId}`); 
      return; 
    }

    initiatePaymentMutation.mutate({
      booking_id: booking.id, 
      amount: booking.total_price,
      payment_method: selectedPaymentMethod,
    });
  };

  const dummyBankAccounts = {
    bca: '123-456-7890',
    bni: '098-765-4321',
    bri: '112-233-4455',
    mandiri: '554-443-2211', 
  };

  const getBankReference = (bookingId: string | number) => {
    const base = String(bookingId);
    if (base.length >= 3) {
      return base.slice(-3); 
    }
    return Math.floor(Math.random() * 900) + 100; 
  };

  return (
    <div className="w-full px-4 max-w-2xl mx-auto bg-card rounded-lg shadow-lg py-8"> {}
      <h2 className="text-2xl font-bold mb-4 text-primary">Payment for Booking ID: {booking.id}</h2> {}
      <p className="text-foreground mb-2"><strong>Concert:</strong> {booking.concert_name}</p> {}
      <p className="text-foreground mb-4"><strong>Total Amount Due:</strong> <span className="font-bold text-primary">{formatCurrency(booking.total_price)}</span></p> {}
      <p className="text-foreground mb-4"><strong>Current Status:</strong> {booking.status}</p> {}

      {(isPaid || isFailed || isCancelled) ? (
        <div className="mt-4 p-4 rounded-md bg-secondary"> {}
          {isPaid && <Alert type="success" title="Payment Confirmed!" description="Your tickets are secured." />}
          {isFailed && <Alert type="error" title="Payment Failed!" description="Please try booking again." />}
          {isCancelled && <Alert type="info" title="Booking Confirmed" description="This booking has been cancelled." />}
          {payment && (
            <div className="mt-4 text-foreground"> {}
              <h4 className="text-lg font-semibold mb-2">Payment Details:</h4>
              <p>Transaction ID: {payment.transaction_id || 'N/A'}</p>
              <p>Method: {payment.payment_method}</p>
              <p>Status: {payment.status}</p>
            </div>
          )}
          <Button onClick={() => navigate('/my-bookings')} className="mt-4 bg-primary hover:bg-primary/90 text-primary-foreground">
            Go to My Bookings
          </Button>
        </div>
      ) : (
        <div className="mt-4">
          {paymentInitiationError && <Alert type="error" description={paymentInitiationError} />}
          
          <h3 className="text-xl font-semibold mb-4 text-foreground">Select Payment Method:</h3> {}
          
          {}
          <Accordion type="single" collapsible className="w-full" onValueChange={(value) => setSelectedPaymentMethod(value)}>
            <AccordionItem value="credit_card">
              <AccordionTrigger className="font-semibold text-lg hover:no-underline text-foreground"> {}
                Credit Card
              </AccordionTrigger>
              <AccordionContent className="p-4 border rounded-b-md bg-card shadow-sm text-foreground"> {}
                <div className="mb-4">
                  <Label htmlFor="cardNumber">Card Number:</Label>
                  <Input 
                    id="cardNumber"
                    type="text" 
                    placeholder="XXXX XXXX XXXX XXXX" 
                    className="mt-1" 
                    value={cardNumber.replace(/(\d{4})(?=\d)/g, '$1 ')} 
                    onChange={handleCardNumberChange}
                    maxLength={19} 
                  />
                  {cardType && (
                    <p className="text-sm text-muted-foreground mt-1">Card Type: <span className="font-semibold text-primary">{cardType}</span></p> 
                  )}
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <Label htmlFor="expiryDate">Expiry Date:</Label>
                    <Input id="expiryDate" type="text" placeholder="MM/YY" className="mt-1" maxLength={5} />
                  </div>
                  <div>
                    <Label htmlFor="cvv">CVV:</Label>
                    <Input id="cvv" type="text" placeholder="XXX" className="mt-1" maxLength={4} />
                  </div>
                </div>
              </AccordionContent>
            </AccordionItem>

            <AccordionItem value="bank_transfer">
              <AccordionTrigger className="font-semibold text-lg hover:no-underline text-foreground"> {}
                Bank Transfer
              </AccordionTrigger>
              <AccordionContent className="p-4 border rounded-b-md bg-card shadow-sm text-foreground"> {}
                <Accordion type="single" collapsible className="w-full">
                  {Object.entries(dummyBankAccounts).map(([bankName, accountNumber]) => (
                    <AccordionItem key={bankName} value={bankName}>
                      <AccordionTrigger className="font-medium text-base hover:no-underline capitalize text-foreground"> {}
                        {bankName.toUpperCase()}
                      </AccordionTrigger>
                      <AccordionContent className="p-4 border rounded-b-md bg-muted text-sm text-muted-foreground"> {}
                        <p className="mb-2"><strong>Account Name:</strong> Concert Ticketing Inc.</p>
                        <p className="mb-2"><strong>Account Number:</strong> {accountNumber}</p>
                        <p className="text-primary font-semibold">Total: {formatCurrency(booking.total_price)}</p> {}
                        <p className="text-sm text-muted-foreground">Please include **{getBankReference(booking.id)}** in your transfer reference.</p>
                      </AccordionContent>
                    </AccordionItem>
                  ))}
                </Accordion>
              </AccordionContent>
            </AccordionItem>

            <AccordionItem value="e_wallet">
              <AccordionTrigger className="font-semibold text-lg hover:no-underline text-foreground"> {}
                E-Wallet (QRIS)
              </AccordionTrigger>
              <AccordionContent className="p-4 border rounded-b-md bg-card shadow-sm text-foreground"> {}
                <p className="text-muted-foreground">Upon clicking "Pay Now", you will be redirected to a page with a QR code to complete your payment.</p> {}
              </AccordionContent>
            </AccordionItem>
          </Accordion>

          <Button 
            onClick={handlePaymentSubmit} 
            disabled={initiatePaymentMutation.isPending || !selectedPaymentMethod}
            className="w-full mt-6 bg-primary hover:bg-primary/90 text-primary-foreground" 
          >
            {initiatePaymentMutation.isPending ? <LoadingSpinner size="small" /> : 'Pay Now'}
          </Button>
        </div>
      )}
    </div>
  );
}

export default PaymentPage;