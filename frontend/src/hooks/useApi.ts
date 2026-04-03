import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import axios from 'axios';
import { getKrakenTraderAPI } from '../client/api';

const apiClient = axios.create({
  baseURL: '/api',
});

const krakenApi = getKrakenTraderAPI(apiClient);

export const useHealth = () =>
  useQuery({
    queryKey: ['health'],
    queryFn: () => krakenApi.getHealth().then(res => res.data),
    refetchInterval: 30000,
  });

export const useSubscriptions = () =>
  useQuery({
    queryKey: ['subscriptions'],
    queryFn: () => krakenApi.listSubscriptions().then(res => res.data),
  });

export const useSubscriptionDetails = () =>
  useQuery({
    queryKey: ['subscriptions', 'detail'],
    queryFn: () => krakenApi.listSubscriptionsDetail().then(res => res.data),
  });

export const useAddSubscription = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (symbol: string) => krakenApi.addSubscription({ symbol }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['subscriptions'] }),
  });
};

export const useRemoveSubscription = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (symbol: string) => krakenApi.deleteSubscription({ symbol }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['subscriptions'] }),
  });
};

export const useAssets = (enabledOnly = true) =>
  useQuery({
    queryKey: ['assets', enabledOnly],
    queryFn: () => krakenApi.getAssets({ enabled_only: enabledOnly }).then(res => res.data),
  });

export const usePairs = () =>
  useQuery({
    queryKey: ['pairs'],
    queryFn: () => krakenApi.getPairs().then(res => res.data),
  });

export const usePrompts = (limit = 20) =>
  useQuery({
    queryKey: ['prompts', limit],
    queryFn: () => krakenApi.listPrompts({ limit }).then(res => res.data),
  });

export const useLogLevel = () =>
  useQuery({
    queryKey: ['loglevel'],
    queryFn: () => krakenApi.getLogLevel().then(res => res.data),
  });

export const useSetLogLevel = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (level: 'trace' | 'debug' | 'info' | 'warn' | 'error') =>
      krakenApi.setLogLevel({ level }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['loglevel'] }),
  });
};
