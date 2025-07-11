import React from 'react';
import { useState } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { getConcertById } from '../api/booking.ts';
import LoadingSpinner from '../components/common/LoadingSpinner';
import Alert from '../components/common/Alert';
import { formatDate, formatCurrency } from '../utils/helpers';
import { CreateBookingSchema } from '../utils/validators';

import { Button } from '../components/ui/button';
import { Input } from '../components/ui/input';

import type { Concert, TicketQuantityByClassRequest } from '../api/booking.ts';

function ConcertDetailPage() {
  const { concertId } = useParams<{ concertId: string }>();
  const navigate = useNavigate();
  const [ticketQuantities, setTicketQuantities] = useState<{ [key: number]: number }>({});

  const { data: concert, isLoading, isError, error } = useQuery<Concert>({
    queryKey: ['concerts', concertId],
    queryFn: () => getConcertById(concertId!),
    staleTime: 5 * 60 * 1000,
    refetchInterval: 10 * 1000,
  });

  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <LoadingSpinner />
      </div>
    );
  }

  if (isError) {
    return <Alert type="error" description={`Failed to load concert details: ${error?.message}`} />;
  }

  if (!concert) {
    return (
      <div className="w-full p-4 max-w-4xl mx-auto bg-card rounded-lg shadow-lg"> {}
        <Alert type="error" description="Concert data not found. It might have been deleted or the ID is invalid." />
        <Link to="/dashboard" className="text-primary hover:underline">Go back to concerts list</Link> {}
      </div>
    );
  }

  if (concert.status !== 'active') {
    return (
      <div className="w-full p-4 max-w-4xl mx-auto bg-card rounded-lg shadow-lg"> {}
        <h2 className="text-2xl font-bold mb-4 text-primary">{concert.name} by {concert.artist}</h2> {}
        {concert.image_url && <img src={concert.image_url} alt={concert.name} className="w-full h-auto rounded-lg mb-6 shadow-lg" />}
        <p className="text-foreground mb-2"><strong>Date:</strong> {formatDate(concert.date)}</p> {}
        <p className="text-foreground mb-2"><strong>Venue:</strong> {concert.venue}</p> {}
        <p className="text-foreground mb-4"><strong>Description:</strong> {concert.description}</p> {}
        <p className="text-primary font-semibold mb-4">This concert is currently {concert.status}. Tickets are not available for booking yet.</p> {}
        <Link to="/dashboard" className="text-primary hover:underline">Go back to concerts list</Link> {}
      </div>
    );
  }

  const handleQuantityChange = (ticketClassId: number, newQuantity: number) => {
    setTicketQuantities(prev => ({
      ...prev,
      [ticketClassId]: newQuantity < 0 ? 0 : newQuantity,
    }));
  };

  const calculateTotalPrice = () => {
    let total = 0;
    if (concert.ticket_classes) {
      for (const tc of concert.ticket_classes) {
        total += (ticketQuantities[tc.id] || 0) * tc.price;
      }
    }
    return total.toFixed(2);
  };

  const totalSelectedTickets = Object.values(ticketQuantities).reduce((sum, qty) => sum + (qty || 0), 0);
  const MAX_TOTAL_TICKETS = 5;

  const handleProceedToBuyerInfo = async () => {
    const selectedTickets: TicketQuantityByClassRequest[] = concert.ticket_classes ? concert.ticket_classes.map(tc => ({
      ticket_class_id: tc.id,
      quantity: ticketQuantities[tc.id] || 0,
    })).filter(tc => tc.quantity > 0) : [];

    if (selectedTickets.length === 0) {
      alert('Please select at least one ticket.');
      return;
    }

    try {
        await CreateBookingSchema.validate(
            {
                concertId: parseInt(concertId!),
                ticketsByClass: selectedTickets.map(ticket => ({
                    ticketClassId: ticket.ticket_class_id, 
                    quantity: ticket.quantity
                })),
            },
            { abortEarly: false }
        );
        
        sessionStorage.setItem('tempBookingTickets', JSON.stringify(selectedTickets));
        sessionStorage.setItem('tempBookingConcert', JSON.stringify({
          id: concert.id,
          name: concert.name,
          date: concert.date,
          totalPrice: parseFloat(calculateTotalPrice()),
        }));
        navigate(`/bookings/confirm/${concertId}`);
    } catch (validationError: any) {
        if (validationError.inner) {
            const messages = validationError.inner.map((err: any) => err.message).join('\n');
            alert(`Validation Error:\n${messages}`);
        } else {
            alert(`Validation Error: ${validationError.message}`);
        }
    }
  };

  return (
    <div className="w-full p-4 max-w-4xl mx-auto bg-card rounded-lg shadow-lg"> {}
      {concert.image_url && <img src={concert.image_url} alt={concert.name} className="w-full h-auto rounded-lg mb-6 shadow-lg" />}
      <h2 className="text-2xl font-bold mb-4 text-primary">{concert.name} by {concert.artist}</h2> {}
      
      <p className="text-foreground mb-2"><strong>Date:</strong> {formatDate(concert.date)}</p> {}
      <p className="text-foreground mb-2"><strong>Venue:</strong> {concert.venue}</p> {}
      <p className="text-foreground mb-4"><strong>Description:</strong> {concert.description}</p> {}
      <p className="text-foreground mb-4"><strong>Total Seats:</strong> {concert.total_seats || 0} / Available: {concert.available_seats || 0}</p> {}

      <h3 className="text-xl font-semibold mb-4 text-foreground">Select Your Tickets (Max {MAX_TOTAL_TICKETS} Tickets Total)</h3> {}
      
      <div className="ticket-class-selection">
        {concert.ticket_classes && concert.ticket_classes.length > 0 ? (
          concert.ticket_classes.map((tc) => (
            <div key={tc.id} className="ticket-class-item p-4 border border-border rounded-md mb-3 bg-background"> {}
              <h4 className="text-lg font-semibold mb-2 text-foreground">{tc.name} - {formatCurrency(tc.price)}</h4> {}
              <p className="text-muted-foreground mb-2">Available: {tc.available_seats_in_class || 0}</p> {}
              <div className="quantity-control flex items-center space-x-2">
                <Button
                  variant="outline" 
                  size="icon"
                  onClick={() => handleQuantityChange(tc.id, (ticketQuantities[tc.id] || 0) - 1)}
                  disabled={(ticketQuantities[tc.id] || 0) === 0}
                  className="bg-secondary text-secondary-foreground hover:bg-secondary/80" 
                >-</Button>
                <Input
                  type="number"
                  min="0"
                  max={tc.available_seats_in_class || 0}
                  value={ticketQuantities[tc.id] || 0}
                  onChange={(e) => {
                    const newQty = parseInt(e.target.value, 10) || 0;
                    handleQuantityChange(tc.id, newQty);
                  }}
                  className="w-20 text-center text-foreground" 
                />
                <Button
                  variant="outline" 
                  size="icon"
                  onClick={() => handleQuantityChange(tc.id, (ticketQuantities[tc.id] || 0) + 1)}
                  disabled={
                    (ticketQuantities[tc.id] || 0) >= (tc.available_seats_in_class || 0) ||
                    totalSelectedTickets >= MAX_TOTAL_TICKETS
                  }
                  className={`bg-secondary text-secondary-foreground hover:bg-secondary/80 ${
                    totalSelectedTickets >= MAX_TOTAL_TICKETS && (ticketQuantities[tc.id] || 0) < (tc.available_seats_in_class || 0) ? 'opacity-50 cursor-not-allowed' : ''
                  }`} 
                >+</Button>
              </div>
            </div>
          ))
        ) : (
          <p className="text-muted-foreground">No ticket classes available for this concert.</p>
        )}
      </div>

      <h3 className="text-xl font-semibold mt-6 text-foreground">Total Price: <span className="text-primary">{formatCurrency(parseFloat(calculateTotalPrice()))}</span></h3> {}
      <Button
        onClick={handleProceedToBuyerInfo}
        disabled={calculateTotalPrice() === '0.00' || totalSelectedTickets === 0 || totalSelectedTickets > MAX_TOTAL_TICKETS}
        className="mt-6 w-full bg-primary hover:bg-primary/90 text-primary-foreground" 
      >
        Proceed to Buyer Info
      </Button>
    </div>
  );
}

export default ConcertDetailPage;