import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import type { AuthModel } from "@/api/types.gen";
import {
  getUsersOptions,
  getUsersQueryKey,
  deleteUsersByIdMutation,
  patchUsersByIdActiveMutation,
  postUsersMutation,
  putUsersByIdPasswordMutation,
  getSettingsRegistrationOptions,
  putSettingsRegistrationMutation,
} from "@/api/@tanstack/react-query.gen";
import Layout from "@/layout";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Switch } from "@/components/ui/switch";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import { toast } from "sonner";
import { Loader2, Trash2, KeyRound } from "lucide-react";
import { commonMutationErrorHandler } from "@/lib/utils";
import { useAuthStore } from "@/store/auth";

const passwordRules = z
  .string()
  .min(8, "Password must be at least 8 characters")
  .regex(/[A-Z]/, "Must contain an uppercase letter")
  .regex(/[a-z]/, "Must contain a lowercase letter")
  .regex(/[0-9]/, "Must contain a number")
  .regex(/[^A-Za-z0-9]/, "Must contain a special character");

const createSchema = z.object({
  email: z.string().email("Please provide a valid email address"),
  password: passwordRules,
});

const resetPasswordSchema = z
  .object({
    password: passwordRules,
    confirm: z.string(),
  })
  .refine((d) => d.password === d.confirm, {
    message: "Passwords do not match",
    path: ["confirm"],
  });

type CreateForm = z.infer<typeof createSchema>;
type ResetPasswordForm = z.infer<typeof resetPasswordSchema>;

const UsersPage = () => {
  const queryClient = useQueryClient();
  const currentUser = useAuthStore((state) => state.user);

  const [showCreate, setShowCreate] = useState(false);
  const [userToDelete, setUserToDelete] = useState<AuthModel | null>(null);
  const [userToResetPassword, setUserToResetPassword] = useState<AuthModel | null>(null);

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<CreateForm>({ resolver: zodResolver(createSchema) });

  const {
    register: registerReset,
    handleSubmit: handleSubmitReset,
    reset: resetPasswordForm,
    formState: { errors: resetErrors },
  } = useForm<ResetPasswordForm>({ resolver: zodResolver(resetPasswordSchema) });

  const { data: usersData, isLoading: usersLoading } = useQuery(getUsersOptions());
  const { data: registrationData } = useQuery(getSettingsRegistrationOptions());

  const users = (usersData?.data as AuthModel[] | undefined) ?? [];
  const registrationEnabled = registrationData?.data?.enabled ?? true;

  const createMutation = useMutation({
    ...postUsersMutation(),
    onSuccess: () => {
      toast.success("User created successfully");
      queryClient.invalidateQueries({ queryKey: getUsersQueryKey() });
      setShowCreate(false);
      reset();
    },
    onError: commonMutationErrorHandler("Failed to create user"),
  });

  const deleteMutation = useMutation({
    ...deleteUsersByIdMutation(),
    onSuccess: () => {
      toast.success("User deleted successfully");
      queryClient.invalidateQueries({ queryKey: getUsersQueryKey() });
      setUserToDelete(null);
    },
    onError: (err) => {
      commonMutationErrorHandler("Failed to delete user")(err);
      setUserToDelete(null);
    },
  });

  const setActiveMutation = useMutation({
    ...patchUsersByIdActiveMutation(),
    onSuccess: () => {
      toast.success("User updated");
      queryClient.invalidateQueries({ queryKey: getUsersQueryKey() });
    },
    onError: commonMutationErrorHandler("Failed to update user"),
  });

  const resetPasswordMutation = useMutation({
    ...putUsersByIdPasswordMutation(),
    onSuccess: () => {
      toast.success("Password reset successfully");
      setUserToResetPassword(null);
      resetPasswordForm();
    },
    onError: commonMutationErrorHandler("Failed to reset password"),
  });

  const registrationMutation = useMutation({
    ...putSettingsRegistrationMutation(),
    onSuccess: (_, vars) => {
      const enabled = (vars.body as { enabled: boolean }).enabled;
      toast.success(enabled ? "Registration enabled" : "Registration disabled");
      queryClient.invalidateQueries({ queryKey: getSettingsRegistrationOptions().queryKey });
    },
    onError: commonMutationErrorHandler("Failed to update registration status"),
  });

  const handleToggleActive = (user: AuthModel, active: boolean) => {
    setActiveMutation.mutate({ path: { id: user.id! }, body: { active } });
  };

  const handleConfirmDelete = () => {
    if (userToDelete?.id) deleteMutation.mutate({ path: { id: userToDelete.id } });
  };

  const onCreateSubmit = (data: CreateForm) => {
    createMutation.mutate({ body: data });
  };

  const onResetPasswordSubmit = (data: ResetPasswordForm) => {
    if (!userToResetPassword?.id) return;
    resetPasswordMutation.mutate({
      path: { id: userToResetPassword.id },
      body: { password: data.password },
    });
  };

  const handleOpenCreate = () => {
    reset();
    setShowCreate(true);
  };

  return (
    <Layout pageName="Users" onCreate={handleOpenCreate}>
      <div className="space-y-6">
        {/* Registration toggle card */}
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-base font-medium">Self-Registration</CardTitle>
          </CardHeader>
          <CardContent className="flex items-center justify-between">
            <span className="text-sm text-muted-foreground">
              Allow new admins to register via the public registration page
            </span>
            <Switch
              checked={registrationEnabled}
              onCheckedChange={(val) => registrationMutation.mutate({ body: { enabled: val } })}
              disabled={registrationMutation.isPending}
            />
          </CardContent>
        </Card>

        {/* Users table */}
        {usersLoading ? (
          <div className="space-y-2">
            {Array.from({ length: 4 }, (_, i) => (
              <Skeleton key={i} className="h-12 w-full rounded-none" />
            ))}
          </div>
        ) : (
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Email</TableHead>
                <TableHead>Status</TableHead>
                <TableHead>2FA</TableHead>
                <TableHead>Created</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {users.map((user) => {
                const isSelf = user.id === currentUser?.id;
                return (
                  <TableRow key={user.id}>
                    <TableCell className="font-medium">{user.email}</TableCell>
                    <TableCell>
                      {user.active ? (
                        <Badge variant="default">Active</Badge>
                      ) : (
                        <Badge variant="secondary">Disabled</Badge>
                      )}
                    </TableCell>
                    <TableCell>
                      {user.twofa_status ? (
                        <Badge variant="default">Enabled</Badge>
                      ) : (
                        <Badge variant="outline">Off</Badge>
                      )}
                    </TableCell>
                    <TableCell className="text-muted-foreground text-sm">
                      {user.createdAt ? new Date(user.createdAt).toLocaleDateString() : "—"}
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex items-center justify-end gap-3">
                        <Switch
                          checked={user.active ?? false}
                          onCheckedChange={(val) => handleToggleActive(user, val)}
                          disabled={isSelf || setActiveMutation.isPending}
                          title={isSelf ? "Cannot disable your own account" : undefined}
                        />
                        <Button
                          variant="ghost"
                          size="icon"
                          title="Reset password"
                          onClick={() => { resetPasswordForm(); setUserToResetPassword(user); }}
                        >
                          <KeyRound className="h-4 w-4" />
                        </Button>
                        <Button
                          variant="ghost"
                          size="icon"
                          className="text-destructive hover:text-destructive"
                          disabled={isSelf || deleteMutation.isPending}
                          title={isSelf ? "Cannot delete your own account" : "Delete user"}
                          onClick={() => setUserToDelete(user)}
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                );
              })}
              {users.length === 0 && (
                <TableRow>
                  <TableCell colSpan={5} className="text-center text-muted-foreground py-8">
                    No users found
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
        )}
      </div>

      {/* Create user modal */}
      <Dialog open={showCreate} onOpenChange={(open) => { if (!open) { setShowCreate(false); reset(); } }}>
        <DialogContent className="sm:max-w-md" onInteractOutside={(e) => e.preventDefault()}>
          <DialogHeader>
            <DialogTitle>Create User</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleSubmit(onCreateSubmit)} className="space-y-4 pt-2">
            <div className="space-y-1">
              <Label htmlFor="email">Email</Label>
              <Input id="email" type="email" {...register("email")} />
              {errors.email && (
                <p className="text-sm text-destructive">{errors.email.message}</p>
              )}
            </div>
            <div className="space-y-1">
              <Label htmlFor="password">Password</Label>
              <Input id="password" type="password" {...register("password")} />
              {errors.password && (
                <p className="text-sm text-destructive">{errors.password.message}</p>
              )}
            </div>
            <div className="flex justify-end gap-2 pt-2">
              <Button type="button" variant="outline" onClick={() => { setShowCreate(false); reset(); }}>
                Cancel
              </Button>
              <Button type="submit" disabled={createMutation.isPending}>
                {createMutation.isPending && <Loader2 className="animate-spin mr-2 h-4 w-4" />}
                Create User
              </Button>
            </div>
          </form>
        </DialogContent>
      </Dialog>

      {/* Reset password modal */}
      <Dialog
        open={!!userToResetPassword}
        onOpenChange={(open) => { if (!open) { setUserToResetPassword(null); resetPasswordForm(); } }}
      >
        <DialogContent className="sm:max-w-md" onInteractOutside={(e) => e.preventDefault()}>
          <DialogHeader>
            <DialogTitle>Reset Password — {userToResetPassword?.email}</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleSubmitReset(onResetPasswordSubmit)} className="space-y-4 pt-2">
            <div className="space-y-1">
              <Label htmlFor="reset-password">New Password</Label>
              <Input id="reset-password" type="password" {...registerReset("password")} />
              {resetErrors.password && (
                <p className="text-sm text-destructive">{resetErrors.password.message}</p>
              )}
            </div>
            <div className="space-y-1">
              <Label htmlFor="reset-confirm">Confirm Password</Label>
              <Input id="reset-confirm" type="password" {...registerReset("confirm")} />
              {resetErrors.confirm && (
                <p className="text-sm text-destructive">{resetErrors.confirm.message}</p>
              )}
            </div>
            <div className="flex justify-end gap-2 pt-2">
              <Button
                type="button"
                variant="outline"
                onClick={() => { setUserToResetPassword(null); resetPasswordForm(); }}
              >
                Cancel
              </Button>
              <Button type="submit" disabled={resetPasswordMutation.isPending}>
                {resetPasswordMutation.isPending && <Loader2 className="animate-spin mr-2 h-4 w-4" />}
                Reset Password
              </Button>
            </div>
          </form>
        </DialogContent>
      </Dialog>

      {/* Delete confirmation */}
      <AlertDialog open={!!userToDelete} onOpenChange={(open) => !open && setUserToDelete(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete User</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete <strong>{userToDelete?.email}</strong>? This action
              cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel onClick={() => setUserToDelete(null)}>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleConfirmDelete}
              disabled={deleteMutation.isPending}
              className="bg-red-600 hover:bg-red-700 focus:ring-red-600"
            >
              {deleteMutation.isPending && <Loader2 className="animate-spin mr-2 h-4 w-4" />}
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </Layout>
  );
};

export default UsersPage;
