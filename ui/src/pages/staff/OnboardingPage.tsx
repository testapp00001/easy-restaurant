import { zodResolver } from '@hookform/resolvers/zod';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { useState } from 'react';
import apiClient from '@/api/client';

import { Button } from '@/components/ui/button';
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { toast } from 'sonner';

// Schema for creating an account
const accountFormSchema = z.object({
  phone_number: z.string().min(10, 'Phone number is required.'),
  name: z.string().optional(),
});

const sessionFormSchema = z.object({
  phone_number: z.string().min(10, 'Phone number is required.'),

  // Use z.preprocess to handle type conversion explicitly
  buffet_bundle_id: z
    .number('Bundle ID must be number.')
    .min(1, 'Bundle ID is required.'),
});

export function OnboardingPage() {
  const [newPin, setNewPin] = useState<string | null>(null);

  const accountForm = useForm<z.infer<typeof accountFormSchema>>({
    resolver: zodResolver(accountFormSchema),
    defaultValues: { phone_number: '', name: '' },
  });

  const sessionForm = useForm<z.infer<typeof sessionFormSchema>>({
    resolver: zodResolver(sessionFormSchema),
    defaultValues: { phone_number: '', buffet_bundle_id: 1 },
  });

  async function onAccountSubmit(values: z.infer<typeof accountFormSchema>) {
    try {
      const response = await apiClient.post('/staff/customer-accounts', values);
      setNewPin(response.data.pin);
      toast.success('Customer account created successfully!', {
        description: `Phone: ${response.data.phone_number}, PIN: ${response.data.pin}`,
      });
      sessionForm.setValue('phone_number', values.phone_number);
    } catch (error) {
      toast.error('Failed to create customer account.');
      console.log(error);
    }
  }

  async function onSessionSubmit(values: z.infer<typeof sessionFormSchema>) {
    try {
      const response = await apiClient.post('/staff/customer-sessions', values);
      toast.success('Customer session started successfully!', {
        description: `Session ID: ${response.data.user_session_id}`,
        duration: Infinity,
      });
    } catch (error) {
      toast.error('Failed to start session.');
      console.log(error);
    }
  }

  return (
    <div className="grid gap-6 md:grid-cols-2">
      <Card>
        <CardHeader>
          <CardTitle>Step 1: Create Customer Account</CardTitle>
          <CardDescription>
            Create a new account for a customer and generate their PIN.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Form {...accountForm}>
            <form
              onSubmit={accountForm.handleSubmit(onAccountSubmit)}
              className="space-y-4"
            >
              <FormField
                control={accountForm.control}
                name="phone_number"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Phone Number</FormLabel>
                    <FormControl>
                      <Input placeholder="e.g., 5551234567" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={accountForm.control}
                name="name"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Name (Optional)</FormLabel>
                    <FormControl>
                      <Input placeholder="John Doe" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <Button type="submit">Create Account</Button>
            </form>
          </Form>
          {newPin && (
            <div className="mt-4 rounded-lg border bg-card p-4 text-center">
              <p className="text-sm text-muted-foreground">
                Generated PIN (give to customer):
              </p>
              <p className="text-2xl font-bold tracking-widest">{newPin}</p>
            </div>
          )}
        </CardContent>
      </Card>
      <Card>
        <CardHeader>
          <CardTitle>Step 2: Start Dining Session</CardTitle>
          <CardDescription>
            Start a new session for an existing customer.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Form {...sessionForm}>
            <form
              onSubmit={sessionForm.handleSubmit(onSessionSubmit)}
              className="space-y-4"
            >
              <FormField
                control={sessionForm.control}
                name="phone_number"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Phone Number</FormLabel>
                    <FormControl>
                      <Input placeholder="Customer's phone number" {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={sessionForm.control}
                name="buffet_bundle_id"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Buffet Bundle ID</FormLabel>
                    <FormControl>
                      <Input
                        type="number"
                        placeholder="e.g., 1 for Basic, 2 for Premium"
                        {...field}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <Button type="submit">Start Session</Button>
            </form>
          </Form>
        </CardContent>
      </Card>
    </div>
  );
}
