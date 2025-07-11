
import * as Yup from 'yup';


export const LoginSchema = Yup.object().shape({
  username: Yup.string().required('Username is required'),
  password: Yup.string().required('Password is required'),
});


export const RegisterSchema = Yup.object().shape({
  username: Yup.string()
    .min(3, 'Username must be at least 3 characters')
    .max(50, 'Username must not exceed 50 characters')
    .required('Username is required'),
  email: Yup.string().email('Invalid email address').required('Email is required'),
  password: Yup.string()
    .min(6, 'Password must be at least 6 characters')
    .required('Password is required'),
  confirmPassword: Yup.string()
    .oneOf([Yup.ref('password')], 'Passwords must match')
    .required('Confirm password is required'),
});


export const CreateBookingSchema = Yup.object().shape({
  concertId: Yup.number().required('Concert ID is required').positive('Concert ID must be positive'),
  ticketsByClass: Yup.array()
    .of(
      Yup.object().shape({
        ticketClassId: Yup.number().required('Ticket class ID is required').positive('Ticket class ID must be positive'),
        quantity: Yup.number().required('Quantity is required').min(1, 'Quantity must be at least 1').integer('Quantity must be an integer'),
      })
    )
    .min(1, 'Please select at least one ticket')
    .required('Ticket selection is required')
    .test(
      'total-quantity-max',
      'You can book a maximum of 5 tickets in total.',
      function (value) {
        const totalQuantity = value ? value.reduce((sum, item) => sum + (item.quantity || 0), 0) : 0;
        return totalQuantity <= 5;
      }
    ),
});


export const BuyerInfoSchema = Yup.object().shape({
  fullName: Yup.string().required('Full Name is required'),
  phoneNumber: Yup.string()
    .matches(/^[0-9]{10,15}$/, 'Phone number is not valid')
    .required('Phone Number is required'),
  email: Yup.string().email('Invalid email address').required('Email is required'),
  ktpNumber: Yup.string()
    .matches(/^[0-9]{16}$/, 'KTP Number must be 16 digits')
    .required('KTP Number is required'),
  buyForSomeoneElse: Yup.boolean().optional(),
  ticketHolderInfo: Yup.object().when('buyForSomeoneElse', {
    is: true,
    then: (schema) =>
      schema.shape({
        fullName: Yup.string().required('Ticket Holder Name is required'),
        ktpNumber: Yup.string()
          .matches(/^[0-9]{16}$/, 'Ticket Holder KTP Number must be 16 digits')
          .required('Ticket Holder KTP Number is required'),
      }),
    otherwise: (schema) => schema.optional(),
  }),
});


export const CreateConcertSchema = Yup.object().shape({
  name: Yup.string().required('Concert name is required').min(3, 'Concert name must be at least 3 characters'),
  artist: Yup.string().required('Artist name is required'),
  date: Yup.date().required('Date is required').min(new Date(), 'Concert date must be in the future'),
  venue: Yup.string().required('Venue is required'),
  description: Yup.string().optional(),
  imageUrl: Yup.string().url('Invalid image URL').required('Image URL is required'),
  ticketClasses: Yup.array()
    .of(
      Yup.object().shape({
        name: Yup.string().required('Class name is required').min(2, 'Class name must be at least 2 characters'),
        price: Yup.number().required('Price is required').positive('Price must be greater than 0'),
        totalSeatsInClass: Yup.number().required('Total seats in class is required').min(1, 'Total seats in class must be at least 1').integer('Total seats must be an integer'),
      })
    )
    .min(1, 'At least one ticket class is required.')
    .required('Ticket classes are required.'),
});


export type LoginFormType = Yup.InferType<typeof LoginSchema>;
export type RegisterFormType = Yup.InferType<typeof RegisterSchema>;
export type CreateBookingFormType = Yup.InferType<typeof CreateBookingSchema>;
export type BuyerInfoFormType = Yup.InferType<typeof BuyerInfoSchema>;
export type CreateConcertFormType = Yup.InferType<typeof CreateConcertSchema>;