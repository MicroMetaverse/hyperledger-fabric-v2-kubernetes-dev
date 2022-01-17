import { Body, Controller, Get, Post } from '@nestjs/common';
import { AppService } from './app.service';
import { EnrollAdminDto, FunctionDto, RegisterUserDto } from './dto';

@Controller()
export class AppController {
  constructor(private readonly appService: AppService) {}

  @Get()
  getHello(): string {
    return this.appService.getHello();
  }

  @Post('enroll-admin')
  async enrollAdmin(@Body() admin: EnrollAdminDto): Promise<any> {
    return await this.appService.enrollAdmin(admin);
  }

  @Post('register-user')
  async registerUser(@Body() user: RegisterUserDto): Promise<any> {
    return await this.appService.registerUser(user);
  }

  @Post('set-data')
  async setData(@Body() fun: FunctionDto): Promise<any> {
    return await this.appService.invoke(fun);
  }

  @Post('get-data')
  async getData(@Body() fun: FunctionDto): Promise<any> {
    return await this.appService.query(fun);
  }
  @Post('update-data')
  async updateData(@Body() fun: FunctionDto): Promise<any> {
    return await this.appService.invoke(fun);
  }
}
