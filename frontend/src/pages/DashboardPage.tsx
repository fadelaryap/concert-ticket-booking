import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { getAllConcerts } from '../api/booking.ts';
import Alert from '../components/common/Alert';
import { Link } from 'react-router-dom';
import { formatDate, formatCurrency } from '../utils/helpers';
import type { Concert } from '../api/booking.ts';

import { Button } from '../components/ui/button.tsx';

function DashboardPage() {
  const { data: concerts, isLoading, isError, error } = useQuery<Concert[]>({
    queryKey: ['concerts'],
    queryFn: getAllConcerts,
    staleTime: 60 * 1000,
    refetchInterval: 15 * 1000,
  });

  if (isLoading) {
    return (
      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3 p-4"> {}
        {[...Array(3)].map((_, i) => (
          <div key={i} className="bg-card rounded-lg shadow-md p-6 animate-pulse"> {}
            <div className="h-40 bg-muted rounded-md mb-4"></div> {}
            <div className="h-4 bg-muted rounded-md mb-2 w-3/4"></div>
            <div className="h-4 bg-muted rounded-md mb-4 w-1/2"></div>
            <div className="h-24 bg-muted rounded-md mb-4"></div>
            <div className="h-10 bg-muted rounded-md w-full"></div>
          </div>
        ))}
      </div>
    );
  }

  if (isError) {
    return <Alert type="error" description={`Failed to load concerts: ${error.message}`} />;
  }

  if (!concerts) {
    return <Alert type="info" description="No concert data received from the server. Please check back later." />;
  }

  const activeConcerts = concerts.filter(c => c && c.status === 'active');
  const pendingConcerts = concerts.filter(c => c && c.status === 'pending_seat_creation');
  const failedConcerts = concerts.filter(c => c && c.status === 'failed');

  return (
    <div className="w-full p-4">
      <h2 className="text-2xl font-bold mb-4 text-foreground">Upcoming Concerts</h2> {}

      {pendingConcerts.length > 0 && (
        <Alert 
          type="info" 
          title="Concerts In Progress" 
          description={`Some concerts are currently being set up. Please check back shortly: ${pendingConcerts.map(c => c.name).join(', ')}`}
        />
      )}

      {failedConcerts.length > 0 && (
        <Alert 
          type="error" 
          title="Concerts Failed" 
          description={`Some concerts failed to be set up: ${failedConcerts.map(c => c.name).join(', ')}`}
        />
      )}

      {activeConcerts.length > 0 ? (
        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
          {activeConcerts.map((concert) => (
            <div key={concert.id} className="bg-card rounded-lg shadow-md p-6 flex flex-col h-full"> {}
              {concert.image_url && (
                <img src={concert.image_url} alt={concert.name} className="w-full h-40 object-cover rounded-md mb-4" />
              )}
              {}
              <h3 className="text-xl font-semibold mb-2 text-primary hover:underline"> {}
                <Link to={`/concerts/${concert.id}`}>
                  {concert.name} by {concert.artist}
                </Link>
              </h3>
              <p className="text-muted-foreground text-sm mb-1"><strong>Venue:</strong> {concert.venue}</p> {}
              <p className="text-muted-foreground text-sm mb-1"><strong>Date:</strong> {formatDate(concert.date)}</p>
              <p className="text-muted-foreground text-sm mb-4"><strong>Available:</strong> {concert.available_seats || 0} of {concert.total_seats || 0} seats</p>
              
              <h4 className="text-md font-semibold mb-2 text-foreground">Ticket Classes:</h4> {}
              <ul className="list-disc pl-5 text-sm text-muted-foreground mb-4 flex-grow"> {}
                {concert.ticket_classes && concert.ticket_classes.length > 0 ? (
                  concert.ticket_classes.map(tc => (
                    <li key={tc.id}>
                      {tc.name}: {formatCurrency(tc.price)} ({tc.available_seats_in_class || 0} available)
                    </li>
                  ))
                ) : (
                  <li>No ticket classes defined.</li>
                )}
              </ul>

              <Link to={`/concerts/${concert.id}`} className="mt-auto">
                <Button className="w-full bg-primary hover:bg-primary/90 text-primary-foreground">View Details & Book</Button> {}
              </Link>
            </div>
          ))}
        </div>
      ) : (
        <Alert type="info" description="No active concerts found. Please check back later." />
      )}
    </div>
  );
}

export default DashboardPage;