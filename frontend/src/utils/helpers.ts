


export const formatDate = (dateString: string | Date | undefined): string => {
  if (!dateString) return 'N/A';
  const date = new Date(dateString);
  if (isNaN(date.getTime())) return 'Invalid Date'; 
  const options: Intl.DateTimeFormatOptions = { year: 'numeric', month: 'long', day: 'numeric', hour: '2-digit', minute: '2-digit' };
  return date.toLocaleDateString(undefined, options);
};


export const formatCurrency = (amount: number | null | undefined, currency: string = 'IDR'): string => {
  if (typeof amount !== 'number' || isNaN(amount)) {
    return 'N/A';
  }
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: currency,
    minimumFractionDigits: 2,
  }).format(amount);
};


export const getCardType = (cardNumber: string): string | null => {
if (!cardNumber) return null;


const cleanCardNumber = cardNumber.replace(/\D/g, '');

if (cleanCardNumber.startsWith('4')) {
  return 'Visa';
}
if (cleanCardNumber.startsWith('51') || cleanCardNumber.startsWith('52') || cleanCardNumber.startsWith('53') || cleanCardNumber.startsWith('54') || cleanCardNumber.startsWith('55')) {
  return 'Mastercard';
}





return null; 
};