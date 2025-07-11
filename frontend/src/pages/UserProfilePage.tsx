
import React, { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { getProfile, updateProfile } from '../api/auth';
import useAuth from '../hooks/useAuth';
import LoadingSpinner from '../components/common/LoadingSpinner';
import Alert from '../components/common/Alert';
import { Formik, Form, Field } from 'formik'; 
import * as Yup from 'yup';

import { Button } from '../components/ui/button';
import { Input } from '../components/ui/input';
import { Label } from '../components/ui/label';

interface UserProfileData {
  id: number;
  username: string;
  email: string;
  role: string;
}

interface UserProfileFormValues {
  id: number;
  username: string;
  email: string;
}

const ProfileSchema: Yup.ObjectSchema<UserProfileFormValues> = Yup.object().shape({
  id: Yup.number().required(),
  username: Yup.string().required(),
  email: Yup.string().email('Invalid email address').required('Email is required'),
});

const UserProfilePage: React.FC = () => {
  const { user, isLoadingAuth } = useAuth();
  const queryClient = useQueryClient();
  const [submissionError, setSubmissionError] = useState<string | null>(null);
  const [submissionSuccess, setSubmissionSuccess] = useState<string | null>(null);

  const { data: profile, isLoading: isLoadingProfile, isError: isErrorProfile, error: errorProfile } = useQuery<UserProfileData>({
    queryKey: ['userProfile'],
    queryFn: getProfile,
    enabled: !!user,
    staleTime: Infinity,
  });

  const updateProfileMutation = useMutation({
    mutationFn: updateProfile,
    onSuccess: () => {
      setSubmissionSuccess('Profile updated successfully!');
      setSubmissionError(null);
      queryClient.invalidateQueries({ queryKey: ['userProfile'] });
    },
    onError: (err: any) => {
      setSubmissionSuccess(null);
      setSubmissionError(err.response?.data?.error || err.message || 'Failed to update profile.');
      console.error('Failed to update profile:', err);
    },
  });

  if (isLoadingAuth || isLoadingProfile) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <LoadingSpinner />
      </div>
    );
  }

  if (isErrorProfile) {
    return <Alert type="error" description={`Failed to load profile: ${errorProfile.message}`} />;
  }

  if (!profile) {
    return <Alert type="info" description="User profile not found. Please log in again." />;
  }

  const initialValues: UserProfileFormValues = {
    id: profile.id,
    username: profile.username,
    email: profile.email,
  };

  const handleSubmit = async (values: UserProfileFormValues) => {
    updateProfileMutation.mutate(values);
  };

  return (
    <div className="w-full p-4 max-w-xl mx-auto bg-white rounded-lg shadow-lg">
      <h2 className="text-2xl font-bold mb-4 text-primary">User Profile</h2> {}
      {submissionError && <Alert type="error" description={submissionError} />}
      {submissionSuccess && <Alert type="success" description={submissionSuccess} />}

      <Formik
        initialValues={initialValues}
        validationSchema={ProfileSchema}
        onSubmit={handleSubmit}
        enableReinitialize={true}
      >
        {({ errors, touched, isSubmitting }) => (
          <Form className="space-y-4">
            <div className="form-group">
              <Label htmlFor="id">User ID:</Label>
              <Field
                name="id"
                type="text"
                as={Input}
                disabled
                className="bg-muted text-muted-foreground cursor-not-allowed" 
              />
            </div>

            <div className="form-group">
              <Label htmlFor="username">Username:</Label>
              <Field
                name="username"
                type="text"
                as={Input}
                disabled
                className="bg-muted text-muted-foreground cursor-not-allowed" 
              />
              <p className="text-xs text-muted-foreground mt-1">Username cannot be changed.</p> {}
            </div>

            <div className="form-group">
              <Label htmlFor="email">Email:</Label>
              <Field
                name="email"
                type="email"
                as={Input}
                className={errors.email && touched.email ? 'border-destructive text-destructive' : ''} 
                disabled={isSubmitting}
              />
              {errors.email && touched.email && (
                <p className="text-destructive text-sm mt-1">{errors.email}</p> 
              )}
            </div>
            
            <Button type="submit" disabled={isSubmitting || updateProfileMutation.isPending} className="w-full bg-primary hover:bg-primary/90 text-primary-foreground">
              {updateProfileMutation.isPending ? <LoadingSpinner size="small" /> : 'Update Profile'}
            </Button>
          </Form>
        )}
      </Formik>
    </div>
  );
};

export default UserProfilePage;