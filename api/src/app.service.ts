import { Injectable } from '@nestjs/common';
import { EnrollAdminDto, FunctionDto, RegisterUserDto } from './dto';
import { enrollAdmin, invoke, query, registerUser } from './fabric';

@Injectable()
export class AppService {
  getHello(): string {
    return 'Explore Swagger UI at <a href="swagger">/swagger</a>';
  }

  async enrollAdmin(admin: EnrollAdminDto): Promise<any> {
    return await enrollAdmin(admin);
  }

  async registerUser(user: RegisterUserDto): Promise<any> {
    return await registerUser(user);
  }

  async query(fun: FunctionDto): Promise<any> {
    return await query(fun);
  }

  async invoke(fun: FunctionDto): Promise<any> {
    return await invoke(fun);
  }
}
