
import React from 'react';
import { Formik, Form, Field, ErrorMessage as FormikErrorMessage } from 'formik';
import { LoginSchema } from '../../utils/validators';
import LoadingSpinner from '../common/LoadingSpinner';


import { Button } from '../ui/button';
import { Input } from '../ui/input';
import { Label } from '../ui/label';

interface LoginFormProps {
  onSubmit: (values: any) => void;
  isLoading: boolean;
}

const LoginForm: React.FC<LoginFormProps> = ({ onSubmit, isLoading }) => {
  return (
    <Formik
      initialValues={{ username: '', password: '' }}
      validationSchema={LoginSchema}
      onSubmit={onSubmit}
    >
      {({ errors, touched }) => (
        <Form className="space-y-4"> {}
          <div className="form-group">
            <Label htmlFor="username">Username:</Label>
            <Field
              id="username"
              name="username"
              type="text"
              as={Input}
              className={errors.username && touched.username ? 'border-red-500' : ''}
              disabled={isLoading}
            />
            <FormikErrorMessage name="username" component="div" className="text-red-500 text-sm mt-1" />
          </div>

          <div className="form-group">
            <Label htmlFor="password">Password:</Label>
            <Field
              id="password"
              name="password"
              type="password"
              as={Input}
              className={errors.password && touched.password ? 'border-red-500' : ''}
              disabled={isLoading}
            />
            <FormikErrorMessage name="password" component="div" className="text-red-500 text-sm mt-1" />
          </div>

          <Button type="submit" disabled={isLoading} className="w-full bg-blue-600 hover:bg-blue-700 text-white">
            {isLoading ? <LoadingSpinner size="small" /> : 'Login'}
          </Button>
        </Form>
      )}
    </Formik>
  );
};

export default LoginForm;