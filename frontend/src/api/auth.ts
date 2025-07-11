
import api from './index';


export interface LoginCredentials {
  username?: string;
  password?: string;
}

export interface RegisterUserData {
  username?: string;
  email?: string;
  password?: string;
  confirmPassword?: string;
}

export interface UserProfile {
  id: number;
  username: string;
  email: string;
  role: string;
  created_at: string; 
  updated_at: string; 
  last_login: string | null; 
}


export interface LoginResponse {
  message: string;
  user: UserProfile;
}

export interface RegisterResponse extends UserProfile {}
export interface GetProfileResponse extends UserProfile {}
export interface UpdateProfileResponse extends UserProfile {}


const USER_SERVICE_PATH = '/users'; 

export const login = async (credentials: LoginCredentials): Promise<LoginResponse> => {
  const response = await api.post(`${USER_SERVICE_PATH}/login`, credentials);
  return response.data;
};

export const register = async (userData: RegisterUserData): Promise<RegisterResponse> => {
  const response = await api.post(`${USER_SERVICE_PATH}/register`, userData);
  return response.data;
};

export const getProfile = async (): Promise<GetProfileResponse> => {
  const response = await api.get(`${USER_SERVICE_PATH}/profile`);
  return response.data;
};

export const updateProfile = async (profileData: Partial<UserProfile>): Promise<UpdateProfileResponse> => {
  const response = await api.put(`${USER_SERVICE_PATH}/profile`, profileData);
  return response.data;
};


export const logout = async (): Promise<{ message: string }> => {
  const response = await api.post(`${USER_SERVICE_PATH}/logout`);
  return response.data;
};