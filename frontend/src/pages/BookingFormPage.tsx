
import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { createBooking } from '../api/booking';
import { BuyerInfoSchema } from '../utils/validators';
import LoadingSpinner from '../components/common/LoadingSpinner';
import Alert from '../components/common/Alert';
import { Formik, Form, Field, ErrorMessage as FormikErrorMessage } from 'formik';
import { formatDate, formatCurrency } from '../utils/helpers';

import { Button } from '../components/ui/button';
import { Input } from '../components/ui/input';
import { Label } from '../components/ui/label';
import { Checkbox } from '../components/ui/checkbox';



interface TempTicketData {
  ticket_class_id: number; 
  quantity: number;
}

interface TempConcertInfo {
  id: number;
  name: string;
  date: string;
  totalPrice: number;
}

interface ApiError extends Error {
  response?: {
    data?: {
      error?: string;
    };
  };
}

function BookingFormPage() {
  const { concertId } = useParams<{ concertId: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [tempBookingData, setTempBookingData] = useState<TempTicketData[] | null>(null);
  const [concertInfo, setConcertInfo] = useState<TempConcertInfo | null>(null);
  const [submissionError, setSubmissionError] = useState<string | null>(null);

  useEffect(() => {
    const tickets = sessionStorage.getItem('tempBookingTickets');
    const concert = sessionStorage.getItem('tempBookingConcert');
    if (tickets && concert) {
      setTempBookingData(JSON.parse(tickets));
      setConcertInfo(JSON.parse(concert));
    } else {
      alert('No selected tickets found. Please select tickets first.'); 
      navigate(`/concerts/${concertId || ''}`);
    }
  }, [concertId, navigate]);

  const createBookingMutation = useMutation({
    mutationFn: createBooking,
    onSuccess: (data) => {
      alert('Booking created successfully! Redirecting to payment.'); 
      sessionStorage.removeItem('tempBookingTickets');
      sessionStorage.removeItem('tempBookingConcert');
      
      queryClient.invalidateQueries({ queryKey: ['bookings', 'my'] });
      queryClient.invalidateQueries({ queryKey: ['concerts', concertInfo?.id] });
      
      navigate(`/payment/${data.id}`);
    },
    onError: (error: ApiError) => {
      console.error('Failed to create booking with buyer info:', error.response?.data?.error || error.message);
      setSubmissionError(error.response?.data?.error || 'Failed to submit booking. Please try again.');
    },
  });

  if (!tempBookingData || !concertInfo) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <LoadingSpinner />
      </div>
    );
  }

  const initialValues = {
    fullName: '',
    phoneNumber: '',
    email: '',
    ktpNumber: '',
    buyForSomeoneElse: false,
    ticketHolderInfo: {
      fullName: '',
      ktpNumber: '',
    },
  };

  const handleSubmit = async (values: typeof initialValues) => {
    setSubmissionError(null);
    const fullBookingDetails = {
      concert_id: concertInfo.id,
      tickets_by_class: tempBookingData.map(item => ({
        ticket_class_id: item.ticket_class_id, 
        quantity: item.quantity,
      })),
      buyer_info: {
        full_name: values.fullName,
        phone_number: values.phoneNumber,
        email: values.email,
        ktp_number: values.ktpNumber,
      },
      ticket_holder_info: values.buyForSomeoneElse
        ? {
            full_name: values.ticketHolderInfo.fullName,
            ktp_number: values.ticketHolderInfo.ktpNumber,
          }
        : null,
    };
    
    createBookingMutation.mutate(fullBookingDetails);
  };

  return (
    <div className="container mx-auto p-4 max-w-2xl bg-white rounded-lg shadow-lg">
      <h2 className="text-2xl font-bold mb-4 text-blue-700">Confirm Your Booking for: {concertInfo.name}</h2>
      <p className="text-gray-700 mb-2"><strong>Date:</strong> {formatDate(concertInfo.date)}</p>
      <p className="text-gray-700 mb-4"><strong>Total Price:</strong> {formatCurrency(concertInfo.totalPrice)}</p>
      <p className="text-gray-700 mb-4"><strong>Selected Tickets:</strong></p>
      <ul className="list-disc pl-5 mb-4 text-gray-700">
        {tempBookingData.map((item, index) => (
          <li key={index}>Class ID: {item.ticket_class_id}, Quantity: {item.quantity}</li>
        ))}
      </ul>

      <h3 className="text-xl font-semibold mb-4 text-gray-800">Your Information</h3>
      {submissionError && <Alert type="error" description={submissionError} />}

      <Formik
        initialValues={initialValues}
        validationSchema={BuyerInfoSchema}
        onSubmit={handleSubmit}
      >
        {({ errors, touched, values, setFieldValue }) => (
          <Form className="space-y-4">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div className="form-group">
                <Label htmlFor="fullName">Full Name:</Label>
                <Field
                  id="fullName"
                  name="fullName"
                  type="text"
                  as={Input}
                  className={errors.fullName && touched.fullName ? 'border-red-500' : ''}
                />
                <FormikErrorMessage name="fullName" component="div" className="text-red-500 text-sm mt-1" />
              </div>

              <div className="form-group">
                <Label htmlFor="phoneNumber">Phone Number:</Label>
                <Field
                  id="phoneNumber"
                  name="phoneNumber"
                  type="text"
                  as={Input}
                  className={errors.phoneNumber && touched.phoneNumber ? 'border-red-500' : ''}
                />
                <FormikErrorMessage name="phoneNumber" component="div" className="text-red-500 text-sm mt-1" />
              </div>

              <div className="form-group">
                <Label htmlFor="email">Email:</Label>
                <Field
                  id="email"
                  name="email"
                  type="email"
                  as={Input}
                  className={errors.email && touched.email ? 'border-red-500' : ''}
                />
                <FormikErrorMessage name="email" component="div" className="text-red-500 text-sm mt-1" />
              </div>

              <div className="form-group">
                <Label htmlFor="ktpNumber">KTP Number:</Label>
                <Field
                  id="ktpNumber"
                  name="ktpNumber"
                  type="text"
                  as={Input}
                  className={errors.ktpNumber && touched.ktpNumber ? 'border-red-500' : ''}
                />
                <FormikErrorMessage name="ktpNumber" component="div" className="text-red-500 text-sm mt-1" />
              </div>
            </div>

            <div className="flex items-center space-x-2 mt-4">
              <Checkbox
                id="buyForSomeoneElse"
                checked={values.buyForSomeoneElse}
                onCheckedChange={(checked: boolean) => setFieldValue('buyForSomeoneElse', checked)}
              />
              <Label htmlFor="buyForSomeoneElse">Buy for someone else?</Label>
            </div>

            {values.buyForSomeoneElse && (
              <div className="mt-4 border-t pt-4 border-gray-200">
                <h3 className="text-lg font-semibold mb-3 text-gray-800">Ticket Holder Information</h3>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div className="form-group">
                    <Label htmlFor="ticketHolderInfo.fullName">Ticket Holder's Full Name:</Label>
                    <Field
                      id="ticketHolderInfo.fullName"
                      name="ticketHolderInfo.fullName"
                      type="text"
                      as={Input}
                      className={errors.ticketHolderInfo?.fullName && touched.ticketHolderInfo?.fullName ? 'border-red-500' : ''}
                    />
                    <FormikErrorMessage name="ticketHolderInfo.fullName" component="div" className="text-red-500 text-sm mt-1" />
                  </div>
                  <div className="form-group">
                    <Label htmlFor="ticketHolderInfo.ktpNumber">Ticket Holder's KTP Number:</Label>
                    <Field
                      id="ticketHolderInfo.ktpNumber"
                      name="ticketHolderInfo.ktpNumber"
                      type="text"
                      as={Input}
                      className={errors.ticketHolderInfo?.ktpNumber && touched.ticketHolderInfo?.ktpNumber ? 'border-red-500' : ''}
                    />
                    <FormikErrorMessage name="ticketHolderInfo.ktpNumber" component="div" className="text-red-500 text-sm mt-1" />
                  </div>
                </div>
              </div>
            )}

            <Button type="submit" disabled={createBookingMutation.isPending} className="w-full bg-blue-600 hover:bg-blue-700 text-white mt-6">
              {createBookingMutation.isPending ? <LoadingSpinner size="small" /> : 'Submit Booking & Go to Payment'}
            </Button>
          </Form>
        )}
      </Formik>
    </div>
  );
}

export default BookingFormPage;