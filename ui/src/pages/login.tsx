import { useState } from "react";
import { Button } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useLogin } from "@/hooks/use-auth";

export function LoginPage() {
	const [email, setEmail] = useState("");
	const [password, setPassword] = useState("");
	const login = useLogin();

	const handleSubmit = async (e: React.FormEvent) => {
		e.preventDefault();
		try {
			await login.mutateAsync({ email, password });
			window.location.href = "/videos";
		} catch (error) {
			console.error("Login failed:", error);
		}
	};

	return (
		<div className="flex min-h-screen items-center justify-center bg-gray-50 px-4">
			<Card className="w-full max-w-md">
				<CardHeader>
					<CardTitle>Login</CardTitle>
					<CardDescription>
						Enter your credentials to access your account
					</CardDescription>
				</CardHeader>
				<CardContent>
					<form onSubmit={handleSubmit} className="space-y-4">
						<div className="space-y-2">
							<Label htmlFor="email">Email</Label>
							<Input
								id="email"
								type="email"
								placeholder="you@example.com"
								value={email}
								onChange={(e) => setEmail(e.target.value)}
								required
							/>
						</div>
						<div className="space-y-2">
							<Label htmlFor="password">Password</Label>
							<Input
								id="password"
								type="password"
								placeholder="••••••••"
								value={password}
								onChange={(e) => setPassword(e.target.value)}
								required
							/>
						</div>
						{login.isError && (
							<div className="text-sm text-red-600">
								{login.error instanceof Error
									? login.error.message
									: "Login failed"}
							</div>
						)}
						<Button type="submit" className="w-full" disabled={login.isPending}>
							{login.isPending ? "Logging in..." : "Login"}
						</Button>
						<div className="text-center text-sm">
							Don't have an account?{" "}
							<a href="/register" className="text-blue-600 hover:underline">
								Register
							</a>
						</div>
					</form>
				</CardContent>
			</Card>
		</div>
	);
}
